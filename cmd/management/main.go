package main

import (
	"context"
	"crypto/tls"
	"expvar"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"go.uber.org/automaxprocs/maxprocs"
	"go.uber.org/zap"

	"github.com/canonical/lxd-cluster-manager/config"
	routes "github.com/canonical/lxd-cluster-manager/internal/app/management/api"
	"github.com/canonical/lxd-cluster-manager/internal/pkg/api"
	"github.com/canonical/lxd-cluster-manager/internal/pkg/database"
	"github.com/canonical/lxd-cluster-manager/internal/pkg/logger"
	"github.com/canonical/lxd-cluster-manager/internal/pkg/middleware"
)

var build = "development"
var service = "MANAGEMENT"

func main() {
	// Construct logger for the service.
	log, err := logger.New(service)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer log.Sync()

	// Perform the startup and shutdown sequence.
	err = run(log)
	if err != nil {
		log.Errorw("startup", "ERROR", err)
		log.Sync()
		os.Exit(1)
	}
}

func run(logger *zap.SugaredLogger) error {

	// =========================================================================
	// GOMAXPROCS

	// Set the correct number of threads for the service
	// based on what is available either by the machine or quotas.
	if _, err := maxprocs.Set(); err != nil {
		return fmt.Errorf("maxprocs: %w", err)
	}
	logger.Infow("startup", "GOMAXPROCS", runtime.GOMAXPROCS(0))

	// =========================================================================
	// Load configuration

	requireCert := true
	cfg, err := config.LoadConfig(requireCert)
	if err != nil {
		logger.Error("Failed to load configuration")
	}

	// =========================================================================
	// App starting

	logger.Infow("starting service", "environment", build)
	expvar.NewString("build").Set(build)
	defer logger.Infow("shutdown complete")

	// =========================================================================
	// Initialize authentication support

	// =========================================================================
	// Database Support

	logger.Infow("startup", "status", "initializing database support", "host", cfg.DBHost)
	dbConfigs := database.DBConfig{
		DBHost:         cfg.DBHost,
		DBUser:         cfg.DBUser,
		DBPassword:     cfg.DBPassword,
		DBName:         cfg.DBName,
		DBMaxIdleConns: cfg.DBMaxIdleConns,
		DBMaxOpenConns: cfg.DBMaxOpenConns,
		DBDisableTLS:   cfg.DBDisableTLS,
		Logger:         logger,
	}

	db, err := database.NewDB(dbConfigs)
	if err != nil {
		return fmt.Errorf("database connection error: %w", err)
	}
	defer func() {
		logger.Infow("shutdown", "status", "stopping database support", "host", cfg.DBHost)
		db.Close()
	}()

	// =========================================================================
	// Initialize api

	logger.Infow("startup", "status", "initializing API")

	// Make a channel to listen for an interrupt or terminate signal from the OS.
	// Use a buffered channel because the signal package requires it.
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	a := api.NewApi(api.ApiConfig{
		Shutdown: shutdown,
		DB:       db,
		Logger:   logger,
	})

	// register global middlewares in order
	m := middleware.NewMiddleware(logger)
	a.UseGlobalMiddleWares(
		m.RequestTrace,
		m.LogRequest,
	)

	// register api routes
	version := "1.0"
	a.RegisterRoutes(routes.APIRoutes, version)

	// Construct a TLS enabled server to service the requests against the mux.
	tlsConfig := &tls.Config{}
	// List of server certs presented during handshake.
	// NOTE: for the management service we do not need to setup mtls, therefore client certificates are not required.
	tlsConfig.Certificates = []tls.Certificate{cfg.ServerCert}
	tlsConfig.MinVersion = tls.VersionTLS13

	server := http.Server{
		Addr:         cfg.ServerHost + ":" + cfg.ManagementPort,
		Handler:      a,
		ReadTimeout:  time.Duration(cfg.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(cfg.IdleTimeout) * time.Second,
		ErrorLog:     zap.NewStdLog(logger.Desugar()),
		TLSConfig:    tlsConfig,
	}

	// Make a channel to listen for errors coming from the listener. Use a
	// buffered channel so the goroutine can exit if we don't collect this error.
	serverErrors := make(chan error, 1)

	// Start the server listening for requests.
	go func() {
		logger.Infow("startup", "status", "api router started", "host", server.Addr)
		serverErrors <- server.ListenAndServeTLS("", "")
	}()

	// =========================================================================
	// Graceful shutdown

	// Blocking main thread unless if a shutdown signal is received or an server error occurs.
	select {
	case err := <-serverErrors:
		return fmt.Errorf("server error: %w", err)

	case sig := <-shutdown:
		logger.Infow("shutdown", "status", "shutdown started", "signal", sig)
		defer logger.Infow("shutdown", "status", "shutdown complete", "signal", sig)

		// Give outstanding requests a deadline for completion.
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(20*time.Second))
		defer cancel()

		// Asking server to shutdown and shed load.
		if err := server.Shutdown(ctx); err != nil {
			server.Close()
			return fmt.Errorf("could not stop server gracefully: %w", err)
		}
	}

	return nil
}

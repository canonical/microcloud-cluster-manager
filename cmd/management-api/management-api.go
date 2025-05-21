package managementapi

import (
	"context"
	"expvar"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	routes "github.com/canonical/lxd-cluster-manager/internal/app/management-api/api"
	"github.com/canonical/lxd-cluster-manager/internal/app/management-api/core/auth"
	"github.com/canonical/lxd-cluster-manager/internal/pkg/api"
	"github.com/canonical/lxd-cluster-manager/internal/pkg/config"
	"github.com/canonical/lxd-cluster-manager/internal/pkg/database"
	"github.com/canonical/lxd-cluster-manager/internal/pkg/logger"
	"github.com/canonical/lxd-cluster-manager/internal/pkg/middleware"
	"github.com/canonical/lxd/lxd/util"
	"go.uber.org/automaxprocs/maxprocs"
	"go.uber.org/zap"
)

// Run will initialise and start the management-api service API.
func Run() (err error) {
	// =========================================================================
	// GOMAXPROCS

	// Set the correct number of threads for the service
	// based on what is available either by the machine or quotas.
	if _, err := maxprocs.Set(); err != nil {
		return fmt.Errorf("maxprocs: %w", err)
	}
	logger.Log.Infow("startup", "GOMAXPROCS", runtime.GOMAXPROCS(0))

	// =========================================================================
	// Load configuration

	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("Failed to load configuration: %w", err)
	}

	err = cfg.LoadCertificates()
	if err != nil {
		return fmt.Errorf("failed to load certificates: %w", err)
	}

	// =========================================================================
	// App starting

	logger.Log.Infow("starting service", "environment", cfg.Version)
	expvar.NewString("build").Set(cfg.Version)
	defer logger.Log.Infow("shutdown complete")

	// =========================================================================
	// Initialize authentication support

	oidcVerifier, err := auth.NewVerifier(
		cfg.OIDCIssuer,
		cfg.OIDCClientID,
		cfg.OIDCAudience,
		cfg.ManagementAPICert,
	)

	if err != nil {
		return fmt.Errorf("oidc verifier error: %w", err)
	}

	// =========================================================================
	// Database Support

	logger.Log.Infow("startup", "status", "initializing database support", "host", cfg.DBHost)
	dbConfigs := database.DBConfig{
		DBHost:         cfg.DBHost,
		DBUser:         cfg.DBUser,
		DBPassword:     cfg.DBPassword,
		DBName:         cfg.DBName,
		DBMaxIdleConns: cfg.DBMaxIdleConns,
		DBMaxOpenConns: cfg.DBMaxOpenConns,
		DBDisableTLS:   cfg.DBDisableTLS,
	}

	db, err := database.NewDB(dbConfigs)
	if err != nil {
		return fmt.Errorf("database connection error: %w", err)
	}
	defer func() {
		logger.Log.Infow("shutdown", "status", "stopping database support", "host", cfg.DBHost)
		err = db.Close()
	}()

	// =========================================================================
	// Initialize api

	logger.Log.Infow("startup", "status", "initializing API")

	// Make a channel to listen for an interrupt or terminate signal from the OS.
	// Use a buffered channel because the signal package requires it.
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	a := api.NewAPI(api.APIConfig{
		Shutdown:  shutdown,
		DB:        db,
		Auth:      oidcVerifier,
		EnvConfig: cfg,
	})

	// register global middlewares in order
	a.Mux().Use(middleware.RequestTrace)
	a.Mux().Use(middleware.LogRequest)

	// register api routes
	a.RegisterRoutes(routes.APIRoutes)

	// Construct a TLS enabled server to service the requests against the mux.
	tlsConfig := util.ServerTLSConfig(cfg.ManagementAPICert)
	server := http.Server{
		Addr:         cfg.ServerHost + ":" + cfg.ServerPort,
		Handler:      a,
		ReadTimeout:  time.Duration(cfg.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(cfg.IdleTimeout) * time.Second,
		ErrorLog:     zap.NewStdLog(logger.Log.Desugar()),
		TLSConfig:    tlsConfig,
	}

	// Make a channel to listen for errors coming from the listener. Use a
	// buffered channel so the goroutine can exit if we don't collect this error.
	serverErrors := make(chan error, 1)

	// Start the server listening for requests.
	go func() {
		logger.Log.Infow("startup", "status", "api router started", "host", server.Addr)
		serverErrors <- server.ListenAndServeTLS("", "")
	}()

	// setup status endpoint on different port
	statusServer := a.GetStatusServer()
	go func() {
		logger.Log.Infow("startup", "status", "status server started", "host", statusServer.Addr)
		serverErrors <- statusServer.ListenAndServe()
	}()

	// =========================================================================
	// Graceful shutdown

	// Blocking main thread unless if a shutdown signal is received or an server error occurs.
	select {
	case err := <-serverErrors:
		return fmt.Errorf("server error: %w", err)

	case sig := <-shutdown:
		logger.Log.Infow("shutdown", "status", "shutdown started", "signal", sig)
		defer logger.Log.Infow("shutdown", "status", "shutdown complete", "signal", sig)

		// Give outstanding requests a deadline for completion.
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(20*time.Second))
		defer cancel()

		// Asking server to shutdown and shed load.
		if err := server.Shutdown(ctx); err != nil {
			err = server.Close()
			return fmt.Errorf("could not stop server gracefully: %w", err)
		}
	}

	return nil
}

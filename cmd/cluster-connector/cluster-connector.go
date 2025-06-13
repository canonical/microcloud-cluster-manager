package clusterconnector

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

	"github.com/canonical/lxd/lxd/util"
	routes "github.com/canonical/microcloud-cluster-manager/internal/app/cluster-connector/api"
	"github.com/canonical/microcloud-cluster-manager/internal/app/cluster-connector/core/auth"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/api"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/config"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/database"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/database/schema"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/database/seed"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/logger"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/middleware"
	"go.uber.org/automaxprocs/maxprocs"
	"go.uber.org/zap"
)

// Run will initialise and start the cluster-connector service API.
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

	// time out the database connection after 5 minutes
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// ensure the database is ready
	err = db.StatusCheck(ctx)
	if err != nil {
		logger.Log.Errorw("startup", "status", "database connection timeout", "ERROR", err)
		return err
	}

	// =========================================================================
	// Migrate the database

	logger.Log.Infow("startup", "message", "Starting database migration")

	applied, err := schema.Migrate(ctx, db.Conn().DB, cfg.Version)
	if applied {
		logger.Log.Infow("startup", "status", "Database version matches the environment version, no migration needed")
	}

	if err != nil {
		logger.Log.Errorw("startup", "status", "Database migration failed", "ERROR", err)
		return err
	}

	logger.Log.Infow("startup", "status", "Finished database migration")

	// =========================================================================
	// Seed the database

	if cfg.Version == "development" {
		logger.Log.Infow("startup", "message", "Starting database seeding")

		err := seed.SeedDatabase(ctx, db)
		if err != nil {
			logger.Log.Errorw("startup", "status", "database seeding failed", "ERROR", err)
			return err
		}

		logger.Log.Infow("startup", "status", "database seeding successful")
	}

	// =========================================================================
	// Initialize authentication support

	mtlsAuthenticator := auth.NewMtlsAuthenticator(db)

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
		EnvConfig: cfg,
		Auth:      mtlsAuthenticator,
	})

	// register global middlewares in order
	a.Mux().Use(middleware.RequestTrace)
	a.Mux().Use(middleware.LogRequest)

	// register api routes
	a.RegisterRoutes(routes.APIRoutes)

	// NOTE: Construct a TLS enabled server to service the requests against the mux.
	tlsConfig := util.ServerTLSConfig(cfg.ClusterConnectorCert)
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
	statusServer := a.GetStatusServer(cfg.StatusPort)
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

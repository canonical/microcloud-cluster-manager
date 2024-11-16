package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/canonical/lxd-cluster-manager/config"
	"github.com/canonical/lxd-cluster-manager/internal/pkg/database"
	"github.com/canonical/lxd-cluster-manager/internal/pkg/database/schema"
	"github.com/canonical/lxd-cluster-manager/internal/pkg/logger"
	"go.uber.org/zap"
)

var service = "admin"

func main() {
	log, err := logger.New(service)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer log.Sync()

	err = migrate(log)
	if err != nil {
		log.Errorw("admin", "ERROR", err)
		log.Sync()
		os.Exit(1)
	}
}

func migrate(log *zap.SugaredLogger) error {
	log.Infow("migrate", "message", "Migrating the database")

	// =========================================================================
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Error("Failed to load configuration")
	}

	// =========================================================================
	// connect to database
	log.Infow("admin migrate", "status", "connecting to the database", "host", cfg.DBHost)
	dbConfigs := database.DBConfig{
		DBHost:         cfg.DBHost,
		DBUser:         cfg.DBUser,
		DBPassword:     cfg.DBPassword,
		DBName:         cfg.DBName,
		DBMaxIdleConns: cfg.DBMaxIdleConns,
		DBMaxOpenConns: cfg.DBMaxOpenConns,
		DBDisableTLS:   cfg.DBDisableTLS,
		Logger:         log,
	}

	db, err := database.NewDB(dbConfigs)
	if err != nil {
		return fmt.Errorf("database connection error: %w", err)
	}
	defer func() {
		log.Infow("shutdown", "status", "stopping database", "host", cfg.DBHost)
		db.Close()
	}()

	// =========================================================================
	// Migrate the database
	// time out the database migration after 5 minutes
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// ensure the database is ready
	err = db.StatusCheck(ctx)
	if err != nil {
		log.Errorw("admin migrate", "status", "database not ready", "ERROR", err)
		return err
	}

	applied, err := schema.Migrate(ctx, db.Conn().DB, cfg.Version)
	if applied {
		log.Infow("admin migrate", "status", "database version matches the environment version, no migration needed")
		return nil
	}

	if err != nil {
		log.Errorw("admin migrate", "status", "database migration failed", "ERROR", err)
		return err
	}

	log.Infow("admin migrate", "status", "database migration successful")
	return nil
}

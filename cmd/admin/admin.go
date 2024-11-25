package admin

import (
	"context"
	"fmt"
	"time"

	"github.com/canonical/lxd-cluster-manager/config"
	"github.com/canonical/lxd-cluster-manager/internal/pkg/database"
	"github.com/canonical/lxd-cluster-manager/internal/pkg/database/schema"
	"github.com/canonical/lxd-cluster-manager/internal/pkg/logger"
)

// Run will execute all the admin jobs
func Run() error {
	return migrate()
}

func migrate() error {
	logger.Log.Infow("migrate", "message", "Migrating the database")

	// =========================================================================
	// Load configuration

	requireCert := false
	cfg, err := config.LoadConfig(requireCert)
	if err != nil {
		logger.Log.Error("Failed to load configuration")
	}

	// =========================================================================
	// connect to database
	logger.Log.Infow("admin migrate", "status", "connecting to the database", "host", cfg.DBHost)
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
		logger.Log.Infow("shutdown", "status", "stopping database", "host", cfg.DBHost)
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
		logger.Log.Errorw("admin migrate", "status", "database not ready", "ERROR", err)
		return err
	}

	applied, err := schema.Migrate(ctx, db.Conn().DB, cfg.Version)
	if applied {
		logger.Log.Infow("admin migrate", "status", "database version matches the environment version, no migration needed")
		return nil
	}

	if err != nil {
		logger.Log.Errorw("admin migrate", "status", "database migration failed", "ERROR", err)
		return err
	}

	logger.Log.Infow("admin migrate", "status", "database migration successful")
	return nil
}

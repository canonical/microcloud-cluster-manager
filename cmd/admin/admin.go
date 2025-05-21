package admin

import (
	"context"
	"fmt"
	"time"

	"github.com/canonical/lxd-cluster-manager/internal/pkg/config"
	"github.com/canonical/lxd-cluster-manager/internal/pkg/database"
	"github.com/canonical/lxd-cluster-manager/internal/pkg/database/schema"
	"github.com/canonical/lxd-cluster-manager/internal/pkg/database/seed"
	"github.com/canonical/lxd-cluster-manager/internal/pkg/logger"
)

// Run will execute all the admin jobs.
func Run() error {
	return admin()
}

func admin() (err error) {
	logger.Log.Infow("admin", "message", "Starting admin jobs")

	// =========================================================================
	// Load configuration

	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("Failed to load configuration: %w", err)
	}

	// =========================================================================
	// connect to database
	logger.Log.Infow("admin", "status", "connecting to the database", "host", cfg.DBHost)
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
		err = db.Close()
	}()

	// time out the database connection after 5 minutes
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// ensure the database is ready
	err = db.StatusCheck(ctx)
	if err != nil {
		logger.Log.Errorw("admin", "status", "database connection timeout", "ERROR", err)
		return err
	}

	// =========================================================================
	// Migrate the database

	logger.Log.Infow("admin", "message", "Starting database migration")

	applied, err := schema.Migrate(ctx, db.Conn().DB, cfg.Version)
	if applied {
		logger.Log.Infow("admin", "status", "database version matches the environment version, no migration needed")
		return nil
	}

	if err != nil {
		logger.Log.Errorw("admin", "status", "database migration failed", "ERROR", err)
		return err
	}

	logger.Log.Infow("admin", "status", "database migration successful")

	// =========================================================================
	// Seed the database

	if cfg.Version == "development" {
		logger.Log.Infow("admin", "message", "Starting database seeding")

		err := seed.SeedDatabase(ctx, db)
		if err != nil {
			logger.Log.Errorw("admin", "status", "database seeding failed", "ERROR", err)
			return err
		}

		logger.Log.Infow("admin", "status", "database seeding successful")
	}

	return nil
}

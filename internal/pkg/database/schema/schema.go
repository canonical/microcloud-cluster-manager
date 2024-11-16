package schema

import (
	"context"
	"database/sql"
	"embed"

	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var migrations embed.FS

// VERSION_MAP is a map of version strings to their corresponding version number.
var VERSION_MAP = map[string]int64{
	"development": -1,
	"0.1.0":       1,
	"0.1.1":       2,
}

// Migrate runs the database migrations.
func Migrate(ctx context.Context, db *sql.DB, version string) (bool, error) {
	// goose creates a goose_db_version table in the database, which stores the current version of the database
	// we should check if the k8s config version matches that of the database
	// if versions match, then we should skip migration and continue
	currentDBVersion, err := goose.GetDBVersionContext(ctx, db)
	if err != nil {
		return false, err
	}

	// TODO: test this!!!
	// for production, we will apply the migrations up to the specified version
	environmentVersion, ok := VERSION_MAP[version]
	if !ok {
		return false, goose.ErrVersionNotFound
	}

	if currentDBVersion == environmentVersion {
		return true, nil
	}

	// get all the migration sql files
	goose.SetBaseFS(migrations)

	if err := goose.SetDialect("postgres"); err != nil {
		return false, err
	}

	// for development, we will apply all migrations
	if version == "development" {
		if err := goose.UpContext(ctx, db, "migrations"); err != nil {
			return false, err
		}
		return false, nil
	}

	// check if we should upgrade or downgrade db version
	shouldUpgrade := currentDBVersion < environmentVersion
	migrateFunc := goose.UpToContext
	if !shouldUpgrade {
		migrateFunc = goose.DownToContext
	}

	if err := migrateFunc(ctx, db, "migrations", environmentVersion); err != nil {
		return false, err
	}

	return false, nil
}

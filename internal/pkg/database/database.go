// Package database provides support for access the database.
package database

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/canonical/microcloud-cluster-manager/internal/pkg/logger"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/request"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // Calls init function.
)

// DBConfig is the required properties to use the database.
type DBConfig struct {
	DBPort         string
	DBUser         string
	DBPassword     string
	DBHost         string
	DBName         string
	DBMaxIdleConns int
	DBMaxOpenConns int
	DBDisableTLS   bool
}

// DB is a wrapper around the sqlx database connection.
type DB struct {
	cfg  DBConfig
	conn *sqlx.DB
}

// maxRetries is the number of times a function will be retried before giving up.
const maxRetries = 3

// NewDB creates a new database connection.
func NewDB(cfg DBConfig) (*DB, error) {
	db, err := Open(cfg)
	if err != nil {
		return nil, err
	}

	return &DB{
		cfg:  cfg,
		conn: db,
	}, nil
}

// Open knows how to open a database connection based on the configuration.
func Open(cfg DBConfig) (*sqlx.DB, error) {
	sslMode := "require"
	if cfg.DBDisableTLS {
		sslMode = "disable"
	}

	q := make(url.Values)
	q.Set("sslmode", sslMode)
	q.Set("timezone", "utc")

	// construct connection string for db
	u := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(cfg.DBUser, cfg.DBPassword),
		Host:     cfg.DBHost,
		Path:     cfg.DBName,
		RawQuery: q.Encode(),
	}

	db, err := sqlx.Open("postgres", u.String())
	if err != nil {
		return nil, err
	}

	// configure connection pool settings
	db.SetMaxIdleConns(cfg.DBMaxIdleConns)
	db.SetMaxOpenConns(cfg.DBMaxOpenConns)

	return db, nil
}

// isRetriableError determines whether an error is retriable.
// Add logic to detect specific transient errors like deadlocks.
func isRetriableError(err error) bool {
	// TODO: what are other retriable errors?
	if strings.Contains(err.Error(), "deadlock") || strings.Contains(err.Error(), "timeout") {
		return true
	}
	return false
}

// StatusCheck returns nil if it can successfully talk to the database. It
// returns a non-nil error otherwise.
func (db *DB) StatusCheck(ctx context.Context) error {
	// First check we can ping the database.
	var pingError error
	for attempts := 1; ; attempts++ {
		pingError = db.conn.Ping()
		if pingError == nil {
			break
		}
		time.Sleep(time.Duration(attempts) * 100 * time.Millisecond)
		if ctx.Err() != nil {
			return ctx.Err()
		}
	}

	// Make sure we didn't timeout or be cancelled.
	if ctx.Err() != nil {
		return ctx.Err()
	}

	// Run a simple query to determine connectivity. Running this query forces a
	// round trip through the database.
	const q = `SELECT true`
	var tmp bool
	return db.conn.QueryRowContext(ctx, q).Scan(&tmp)
}

// retry executes the provided function with retry logic for transient errors.
// The function fn should return an error if it fails, or nil if it succeeds.
func (db *DB) retry(ctx context.Context, fn func(ctx context.Context) error) error {
	var lastErr error
	traceID := request.GetTraceID(ctx)
	for attempt := 1; attempt <= maxRetries; attempt++ {
		err := fn(ctx)
		if err != nil {
			lastErr = err
			// If the error is not retriable, exit immediately.
			if !isRetriableError(err) {
				return err
			}

			// Log and continue for retriable errors.
			logger.Log.Infow("retry transaction", "traceid", traceID, "attempt", attempt, "maxRetries", maxRetries, "error", err)
			continue
		}
		// If the function succeeds, return nil.
		return nil
	}
	// Return the last error if all retries fail.
	return lastErr
}

// Transaction runs passed function and do commit/rollback at the end.
// Internally it also has a 10s context timeout.
// It will also retry up to 3 times for transient errors.
func (db *DB) Transaction(ctx context.Context, fn func(context.Context, *sqlx.Tx) error) error {
	return db.retry(ctx, func(ctx context.Context) error {
		ctx, cancel := context.WithTimeout(ctx, time.Second*10)
		defer cancel()

		traceID := request.GetTraceID(ctx)

		// Begin the transaction.
		logger.Log.Infow("begin tran", "traceid", traceID)
		tx, err := db.conn.Beginx()
		if err != nil {
			return fmt.Errorf("begin tran: %w", err)
		}

		// Mark to the defer function a rollback is required.
		mustRollback := true

		// Setup a defer function for rolling back the transaction. If
		// mustRollback is true it means the call to fn failed and we
		// need to rollback the transaction.
		defer func() {
			if mustRollback {
				logger.Log.Infow("rollback tran", "traceid", traceID)
				if err := tx.Rollback(); err != nil {
					logger.Log.Errorw("unable to rollback tran", "traceid", traceID, "ERROR", err)
				}
			}
		}()

		// Execute the code inside of the transaction. If the function
		// fails, return the error and the defer function will rollback.
		if err := fn(ctx, tx); err != nil {
			return fmt.Errorf("exec tran: %w", err)
		}

		// Disarm the deferred rollback.
		mustRollback = false

		// Commit the transaction.
		logger.Log.Infow("commit tran", "traceid", traceID)
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit tran: %w", err)
		}

		return nil
	})
}

// Close closes the database connection.
func (db *DB) Close() error {
	return db.conn.Close()
}

// Conn returns the sqlx database connection.
func (db *DB) Conn() *sqlx.DB {
	return db.conn
}

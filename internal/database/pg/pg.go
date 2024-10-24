package pg

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/canonical/lxd/lxd/db/query"
	"github.com/canonical/lxd/shared/logger"
	_ "github.com/lib/pq" // Postgres driver
)

type PgConnection struct {
	DB              *sql.DB
	MaxIdleConns    int
	MaxOpenConns    int
	ConnMaxLifetime time.Duration

	ctx    context.Context
	cancel context.CancelFunc
}

func NewPgConnection(ctx context.Context, host, port, user, password, dbname string, maxIdleConns, maxOpenConns int, connMaxLifetime time.Duration) (*PgConnection, error) {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		logger.Errorf("Failed to connect to Postgres: %v\n", err)
		return nil, fmt.Errorf("failed to connect to Postgres: %w", err)
	}

	// Set connection pool settings
	db.SetMaxIdleConns(maxIdleConns)
	db.SetMaxOpenConns(maxOpenConns)
	db.SetConnMaxLifetime(connMaxLifetime)

	// Ping the database to verify the connection
	if err := db.Ping(); err != nil {
		logger.Errorf("Failed to ping Postgres: %v\n", err)
		return nil, fmt.Errorf("failed to ping Postgres: %w", err)
	}

	shutdownCtx, shutdownCancel := context.WithCancel(ctx)

	logger.Infof("Successfully connected to Postgres at %s:%s\n", host, port)
	return &PgConnection{
		DB:              db,
		MaxIdleConns:    maxIdleConns,
		MaxOpenConns:    maxOpenConns,
		ConnMaxLifetime: connMaxLifetime,
		ctx:             shutdownCtx,
		cancel:          shutdownCancel,
	}, nil
}

func (p *PgConnection) Transaction(ctx context.Context, f func(context.Context, *sql.Tx) error) error {
	// TODO: from what I can tell, lxd query package is a wrapper around the database/sql package, so it should be compatible with the postgres connection here
	// return query.Transaction(ctx, p.DB, f)
	return p.retry(ctx, func(ctx context.Context) error {
		err := query.Transaction(ctx, p.DB, f)
		if errors.Is(err, context.DeadlineExceeded) {
			logger.Warn("Transaction timed out. Retrying once", logger.Ctx{"err": err})
			return query.Transaction(ctx, p.DB, f)
		}

		return err
	})
}

func (p *PgConnection) retry(ctx context.Context, f func(context.Context) error) error {
	if p.ctx.Err() != nil {
		return f(ctx)
	}

	return query.Retry(ctx, f)
}

func (p *PgConnection) Close() error {
	logger.Info("Closing Postgres connection")
	return p.DB.Close()
}

// Bootstrap initializes the PostgreSQL database with the required schema.
func (p *PgConnection) Bootstrap(ctx context.Context) error {
	// Use the Transaction function from the query package to handle the transaction
	return p.Transaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
		// Call the schema update function to execute the schema creation statements
		return pgSchema(ctx, tx)
	})
}

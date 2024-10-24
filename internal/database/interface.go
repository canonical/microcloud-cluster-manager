package database

import (
	"context"
	"database/sql"
)

type Database interface {
	// Transaction handles performing a transaction on the either dqlite or postgres database.
	Transaction(outerCtx context.Context, f func(context.Context, *sql.Tx) error) error

	// TODO: and other methods
}

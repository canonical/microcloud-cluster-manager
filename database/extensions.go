// Package database provides the database access functions and schema.
package database

import (
	"context"
	"database/sql"

	"github.com/canonical/lxd/lxd/db/schema"
)

// SchemaExtensions is a list of schema extensions that can be passed to the MicroCluster daemon.
// Each entry will increase the database schema version by one, and will be applied after internal schema updates.
var SchemaExtensions = []schema.Update{
	schemaAppend1,
}

func schemaAppend1(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, `
CREATE TABLE sites (
    id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
    name TEXT NOT NULL,
    status TEXT NOT NULL,
    UNIQUE (name)
);

CREATE TABLE sites_addresses (
    id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
    site_id INTEGER NOT NULL,
    address TEXT NOT NULL,
    UNIQUE (address),
    FOREIGN KEY (site_id) REFERENCES sites (id)
);
`)
	return err
}

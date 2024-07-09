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
	schemaUpdate1,
}

func schemaUpdate1(ctx context.Context, tx *sql.Tx) error {
	stmt := `
        CREATE TABLE core_sites (
            id                      INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
            name                    TEXT NOT NULL,
            site_certificate        TEXT NOT NULL,
            created_at              DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
            updated_at              DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
            UNIQUE (name),
            UNIQUE (site_certificate)
        );

        CREATE TABLE core_site_tokens (
            id                      INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
            secret                  TEXT NOT NULL,
            expiry                  DATETIME NOT NULL DEFAULT "3000-01-01T00:00:00Z",
            site_name               TEXT NOT NULL,
            created_at              DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
            UNIQUE (site_name)
        );

        CREATE TABLE site_details (
            id                      INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
            core_site_id            INTEGER NOT NULL,
            joined_at               DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
            status                  TEXT NOT NULL CHECK(status IN ('PENDING_APPROVAL', 'ACTIVE')),
            cpu_total_count         FLOAT NOT NULL DEFAULT 0,
            cpu_load_1              TEXT NOT NULL DEFAULT 0,
            cpu_load_5              TEXT NOT NULL DEFAULT 0,
            cpu_load_15             TEXT NOT NULL DEFAULT 0,
            memory_total_amount     FLOAT NOT NULL DEFAULT 0,
            memory_usage            FLOAT NOT NULL DEFAULT 0,
            disk_total_size         FLOAT NOT NULL DEFAULT 0,
            disk_usage              FLOAT NOT NULL DEFAULT 0,
            instance_count          INTEGER NOT NULL DEFAULT 0,
            instance_statuses       TEXT NOT NULL DEFAULT '[]',
            member_count            INTEGER NOT NULL DEFAULT 0,
            member_statuses         TEXT NOT NULL DEFAULT '[]',
            updated_at              DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
            FOREIGN KEY (core_site_id) REFERENCES core_sites (id) ON DELETE CASCADE
        );

        CREATE TABLE manager_configs (
            id                      INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
            key                     TEXT NOT NULL,
            value                   TEXT NOT NULL,
            UNIQUE (key)
        );

        CREATE TABLE manager_member_configs (
            id                      INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
            target                  TEXT NOT NULL,
            https_address           TEXT NOT NULL,
            external_address        TEXT NOT NULL default '',
            FOREIGN KEY (target) REFERENCES internal_cluster_members (name) ON DELETE CASCADE,
            UNIQUE (target)
        );
    `

	_, err := tx.ExecContext(ctx, stmt)
	return err
}

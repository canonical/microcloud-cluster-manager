package pg

import (
	"context"
	"database/sql"
)

func pgSchema(ctx context.Context, tx *sql.Tx) error {
	stmt := `
		CREATE TABLE IF NOT EXISTS core_remote_clusters (
			id SERIAL PRIMARY KEY,
			name TEXT NOT NULL UNIQUE,
			cluster_certificate TEXT NOT NULL UNIQUE,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS core_remote_cluster_tokens (
			id SERIAL PRIMARY KEY,
			secret TEXT NOT NULL,
			expiry TIMESTAMP NOT NULL DEFAULT '3000-01-01 00:00:00',
			cluster_name TEXT NOT NULL UNIQUE,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS remote_cluster_details (
			id SERIAL PRIMARY KEY,
			core_remote_cluster_id INTEGER NOT NULL,
			joined_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			status TEXT NOT NULL CHECK(status IN ('PENDING_APPROVAL', 'ACTIVE')),
			cpu_total_count BIGINT NOT NULL DEFAULT 0,
			cpu_load_1 TEXT NOT NULL DEFAULT '0',
			cpu_load_5 TEXT NOT NULL DEFAULT '0',
			cpu_load_15 TEXT NOT NULL DEFAULT '0',
			memory_total_amount BIGINT NOT NULL DEFAULT 0,
			memory_usage BIGINT NOT NULL DEFAULT 0,
			disk_total_size BIGINT NOT NULL DEFAULT 0,
			disk_usage BIGINT NOT NULL DEFAULT 0,
			instance_count INTEGER NOT NULL DEFAULT 0,
			instance_statuses TEXT NOT NULL DEFAULT '[]',
			member_count INTEGER NOT NULL DEFAULT 0,
			member_statuses TEXT NOT NULL DEFAULT '[]',
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (core_remote_cluster_id) REFERENCES core_remote_clusters (id) ON DELETE CASCADE
		);
    `

	_, err := tx.ExecContext(ctx, stmt)
	return err
}

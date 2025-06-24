package store

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

const DiskThresholdKey = "disk_threshold"
const DiskThresholdDefault = "80" // Default disk threshold percentage
const MemoryThresholdKey = "memory_threshold"
const MemoryThresholdDefault = "80" // Default memory threshold percentage

// RemoteClusterConfig represents a configuration field for a remote cluster.
type RemoteClusterConfig struct {
	ID              int    `db:"id"`                // Primary key
	RemoteClusterID int    `db:"remote_cluster_id"` // Foreign key to remote_clusters
	Key             string `db:"key"`               // Key of the field
	Value           string `db:"value"`             // Value of the field
}

// CreateRemoteClusterConfig creates a new config entry for a remote cluster.
func CreateRemoteClusterConfig(ctx context.Context, tx *sqlx.Tx, data RemoteClusterConfig) (*RemoteClusterConfig, error) {
	q := `
        INSERT INTO remote_cluster_config 
			(remote_cluster_id, key, value)
        VALUES 
			($1, $2, $3)
        RETURNING 
			id, remote_cluster_id, key, value;
    `

	var result RemoteClusterConfig
	err := tx.QueryRowContext(ctx, q,
		data.RemoteClusterID,
		data.Key,
		data.Value,
	).Scan(
		&result.ID,
		&result.RemoteClusterID,
		&result.Key,
		&result.Value,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create \"remote_cluster_config\" entry: %w", err)
	}

	return &result, nil
}

// UpdateRemoteClusterConfig updates a config entry for a remote cluster.
func UpdateRemoteClusterConfig(ctx context.Context, tx *sqlx.Tx, data RemoteClusterConfig) error {
	q := `
	INSERT INTO remote_cluster_config (value, remote_cluster_id, key)
	VALUES ($1, $2, $3)
	ON CONFLICT (remote_cluster_id, key)
	DO UPDATE SET value = EXCLUDED.value;
    `

	result, err := tx.ExecContext(ctx, q,
		data.Value,
		data.RemoteClusterID,
		data.Key,
	)
	if err != nil {
		return fmt.Errorf("update remote_cluster_config entry failed: %w", err)
	}

	n, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("fetch affected rows: %w", err)
	}

	if n != 1 {
		return fmt.Errorf("query updated %d rows instead of 1", n)
	}

	return nil
}

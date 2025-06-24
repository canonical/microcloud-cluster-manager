package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/canonical/lxd/shared/api"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/database/query"
	"github.com/jmoiron/sqlx"
)

// RemoteCluster represents a single remote LXD cluster.
type RemoteCluster struct {
	ID                 int       `db:"id"`                  // Primary key
	Name               string    `db:"name"`                // Cluster name
	Description        string    `db:"description"`         // Description of the cluster (optional)
	Status             string    `db:"status"`              // Status (ACTIVE)
	ClusterCertificate string    `db:"cluster_certificate"` // Unique cluster certificate
	JoinedAt           time.Time `db:"joined_at"`           // Timestamp when joined
	CreatedAt          time.Time `db:"created_at"`          // Creation timestamp
	UpdatedAt          time.Time `db:"updated_at"`          // Update timestamp
}

// GetRemoteClusterID returns the ID of a remote cluster by name.
func GetRemoteClusterID(ctx context.Context, tx *sqlx.Tx, name string) (int, error) {
	// Query to check if the entry exists
	q := `
        SELECT id
        FROM remote_clusters
        WHERE name = $1
    `

	var id int
	err := tx.QueryRowContext(ctx, q, name).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return -1, api.StatusErrorf(http.StatusNotFound, "remote cluster not found")
	}

	if err != nil {
		return -1, fmt.Errorf("failed to get \"remote_clusters\" ID: %w", err)
	}

	return id, nil
}

// RemoteClusterExists checks if a remote cluster with the given key exists.
func RemoteClusterExists(ctx context.Context, tx *sqlx.Tx, name string) (bool, error) {
	_, err := GetRemoteClusterID(ctx, tx, name)
	if err != nil {
		if api.StatusErrorCheck(err, http.StatusNotFound) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

// GetRemoteClusters returns all remote clusters.
func GetRemoteClusters(ctx context.Context, tx *sqlx.Tx) ([]RemoteCluster, error) {
	q := `
        SELECT id, name, description, status, cluster_certificate, joined_at, created_at, updated_at
        FROM remote_clusters;
    `

	objects := make([]RemoteCluster, 0)
	dest := func(scan func(dest ...any) error) error {
		c := RemoteCluster{}
		err := scan(
			&c.ID,
			&c.Name,
			&c.Description,
			&c.Status,
			&c.ClusterCertificate,
			&c.JoinedAt,
			&c.CreatedAt,
			&c.UpdatedAt,
		)
		if err != nil {
			return err
		}

		objects = append(objects, c)

		return nil
	}

	err := query.Scan(ctx, tx, q, dest)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch from \"remote_clusters\" table: %w", err)
	}

	return objects, nil
}

// GetRemoteCluster returns a single remote cluster by name.
func GetRemoteCluster(ctx context.Context, tx *sqlx.Tx, name string) (*RemoteCluster, error) {
	q := `
		SELECT id, name, description, status, cluster_certificate, joined_at, created_at, updated_at
		FROM remote_clusters
		WHERE name = $1;
	`

	var result RemoteCluster
	err := tx.QueryRowContext(ctx, q, name).Scan(
		&result.ID,
		&result.Name,
		&result.Description,
		&result.Status,
		&result.ClusterCertificate,
		&result.JoinedAt,
		&result.CreatedAt,
		&result.UpdatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, api.StatusErrorf(http.StatusNotFound, "remote cluster not found")
	}

	if err != nil {
		return nil, fmt.Errorf("failed to fetch from \"remote_clusters\" table: %w", err)
	}

	return &result, nil
}

// CreateRemoteCluster creates a new remote cluster.
func CreateRemoteCluster(ctx context.Context, tx *sqlx.Tx, data RemoteCluster) (*RemoteCluster, error) {
	exists, err := RemoteClusterExists(ctx, tx, data.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to check for duplicates: %w", err)
	}

	if exists {
		return nil, api.StatusErrorf(http.StatusConflict, "This \"remote_clusters\" entry already exists")
	}

	q := `
        INSERT INTO remote_clusters (name, description, status, cluster_certificate, joined_at, created_at)
        VALUES ($1, $2, $3, $4, $5, $6)
        RETURNING id, name, description, status, cluster_certificate, joined_at, created_at, updated_at;
    `

	var result RemoteCluster
	err = tx.QueryRowContext(ctx, q,
		data.Name,
		data.Description,
		data.Status,
		data.ClusterCertificate,
		data.JoinedAt,
		data.CreatedAt,
	).Scan(
		&result.ID,
		&result.Name,
		&result.Description,
		&result.Status,
		&result.ClusterCertificate,
		&result.JoinedAt,
		&result.CreatedAt,
		&result.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create \"remote_clusters\" entry: %w", err)
	}

	return &result, nil
}

// DeleteRemoteCluster deletes a remote cluster by name.
func DeleteRemoteCluster(ctx context.Context, tx *sqlx.Tx, name string) error {
	q := `
        DELETE FROM remote_clusters
        WHERE name = $1
    `

	result, err := tx.ExecContext(ctx, q, name)
	if err != nil {
		return fmt.Errorf("failed to delete remote cluster: %w", err)
	}

	n, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get number of affected rows: %w", err)
	}

	if n == 0 {
		return api.StatusErrorf(http.StatusNotFound, "no remote cluster found with name: %s", name)
	} else if n > 1 {
		return fmt.Errorf("deleted %d remote clusters instead of 1", n)
	}

	return nil
}

// UpdateRemoteCluster updates a remote cluster by name.
func UpdateRemoteCluster(ctx context.Context, tx *sqlx.Tx, name string, data RemoteCluster) error {
	id, err := GetRemoteClusterID(ctx, tx, name)
	if err != nil {
		return err
	}

	q := `
        UPDATE remote_clusters
        SET name = $1, description = $2, status = $3, joined_at = $4, cluster_certificate = $5
        WHERE id = $6;
    `

	result, err := tx.ExecContext(ctx, q, data.Name, data.Description, data.Status, data.JoinedAt, data.ClusterCertificate, id)
	if err != nil {
		return fmt.Errorf("update remote_clusters entry failed: %w", err)
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

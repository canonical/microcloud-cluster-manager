package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/canonical/lxd-cluster-manager/internal/pkg/database/query"
	"github.com/canonical/lxd/shared/api"
	"github.com/jmoiron/sqlx"
)

// RemoteClusterToken represents a single join token associated with a remote LXD cluster.
type RemoteClusterToken struct {
	ID           int       `db:"id"`            // Primary key
	ClusterName  string    `db:"cluster_name"`  // Name of the associated cluster
	EncodedToken string    `db:"encoded_token"` // EncodedToken token
	Expiry       time.Time `db:"expiry"`        // Expiration timestamp
	CreatedAt    time.Time `db:"created_at"`    // Creation timestamp
}

// GetRemoteClusterTokenID returns the ID of a remote cluster token by name.
func GetRemoteClusterTokenID(ctx context.Context, tx *sqlx.Tx, name string) (int, error) {
	// Query to check if the entry exists
	q := `
        SELECT id
        FROM remote_cluster_tokens
        WHERE cluster_name = $1
    `

	var id int
	err := tx.QueryRowContext(ctx, q, name).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return -1, api.StatusErrorf(http.StatusNotFound, "remote cluster token not found")
	}

	if err != nil {
		return -1, fmt.Errorf("failed to get \"remote_cluster_tokens\" ID: %w", err)
	}

	return id, nil
}

// RemoteClusterTokenExists checks if a remote cluster token with the given key exists.
func RemoteClusterTokenExists(ctx context.Context, tx *sqlx.Tx, name string) (bool, error) {
	_, err := GetRemoteClusterTokenID(ctx, tx, name)
	if err != nil {
		if api.StatusErrorCheck(err, http.StatusNotFound) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

// GetRemoteClusterTokens returns all remote cluster tokens.
func GetRemoteClusterTokens(ctx context.Context, tx *sqlx.Tx) ([]RemoteClusterToken, error) {
	q := `
        SELECT id, cluster_name, encoded_token, expiry, created_at
        FROM remote_cluster_tokens;
    `

	objects := make([]RemoteClusterToken, 0)
	dest := func(scan func(dest ...any) error) error {
		c := RemoteClusterToken{}
		err := scan(
			&c.ID,
			&c.ClusterName,
			&c.EncodedToken,
			&c.Expiry,
			&c.CreatedAt,
		)
		if err != nil {
			return err
		}

		objects = append(objects, c)

		return nil
	}

	err := query.Scan(ctx, tx, q, dest)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch from \"remote_cluster_tokens\" table: %w", err)
	}

	return objects, nil
}

// GetRemoteClusterToken returns a single remote cluster token by name.
func GetRemoteClusterToken(ctx context.Context, tx *sqlx.Tx, name string) (*RemoteClusterToken, error) {
	q := `
		SELECT id, cluster_name, encoded_token, expiry, created_at
        FROM remote_cluster_tokens
		WHERE cluster_name = $1;
	`

	var result RemoteClusterToken
	err := tx.QueryRowContext(ctx, q, name).Scan(
		&result.ID,
		&result.ClusterName,
		&result.EncodedToken,
		&result.Expiry,
		&result.CreatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, api.StatusErrorf(http.StatusNotFound, "remote cluster token not found")
	}

	if err != nil {
		return nil, fmt.Errorf("failed to fetch from \"remote_cluster_tokens\" table: %w", err)
	}

	return &result, nil
}

// CreateRemoteClusterToken creates a new remote cluster token.
func CreateRemoteClusterToken(ctx context.Context, tx *sqlx.Tx, data RemoteClusterToken) (*RemoteClusterToken, error) {
	exists, err := RemoteClusterTokenExists(ctx, tx, data.ClusterName)
	if err != nil {
		return nil, fmt.Errorf("failed to check for duplicates: %w", err)
	}

	if exists {
		return nil, api.StatusErrorf(http.StatusConflict, "this \"remote_cluster_tokens\" entry already exists")
	}

	q := `
        INSERT INTO remote_cluster_tokens (cluster_name, encoded_token, expiry, created_at)
        VALUES ($1, $2, $3, $4)
        RETURNING id, cluster_name, encoded_token, expiry, created_at;
    `

	var result RemoteClusterToken
	err = tx.QueryRowContext(ctx, q,
		data.ClusterName,
		data.EncodedToken,
		data.Expiry,
		data.CreatedAt,
	).Scan(
		&result.ID,
		&result.ClusterName,
		&result.EncodedToken,
		&result.Expiry,
		&result.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create \"remote_cluster_tokens\" entry: %w", err)
	}

	return &result, nil
}

// DeleteRemoteClusterToken deletes a remote cluster token by name.
func DeleteRemoteClusterToken(ctx context.Context, tx *sqlx.Tx, name string) error {
	q := `
        DELETE FROM remote_cluster_tokens
        WHERE cluster_name = $1
    `

	result, err := tx.ExecContext(ctx, q, name)
	if err != nil {
		return fmt.Errorf("failed to delete remote cluster token: %w", err)
	}

	n, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get number of affected rows: %w", err)
	}

	if n == 0 {
		return api.StatusErrorf(http.StatusNotFound, "no remote cluster token found with name: %s", name)
	} else if n > 1 {
		return fmt.Errorf("deleted %d remote cluster tokens instead of 1", n)
	}

	return nil
}

// UpdateCoreRemoteClusterToken updates a remote cluster token by name.
func UpdateCoreRemoteClusterToken(ctx context.Context, tx *sqlx.Tx, name string, data RemoteClusterToken) error {
	id, err := GetRemoteClusterTokenID(ctx, tx, name)
	if err != nil {
		return err
	}

	q := `
        UPDATE remote_cluster_tokens
        SET cluster_name = $1, encoded_token = $2, expiry = $3
        WHERE id = $4;
    `

	result, err := tx.ExecContext(ctx, q, data.ClusterName, data.EncodedToken, data.Expiry, id)
	if err != nil {
		return fmt.Errorf("update remote_cluster_tokens entry failed: %w", err)
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

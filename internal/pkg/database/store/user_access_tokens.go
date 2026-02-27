package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"github.com/canonical/lxd/shared/api"
	"github.com/jmoiron/sqlx"
)

// RemoteClusterToken represents a single join token associated with a remote LXD cluster.
type UserToken struct {
	ID          int    `db:"id"`           // Primary key
	AccessToken string `db:"access_token"` // Name of the associated cluster
}

// GetRemoteClusterTokenID returns the ID of a remote cluster token by name.
func GetUserAccessToken(ctx context.Context, tx *sqlx.Tx, id int) (string, error) {
	// Query to check if the entry exists
	q := `
        SELECT access_token
        FROM user_access_tokens
        WHERE id = $1
    `

	var accessToken string
	err := tx.QueryRowContext(ctx, q, id).Scan(&accessToken)
	if errors.Is(err, sql.ErrNoRows) {
		return "", api.StatusErrorf(http.StatusNotFound, "access token not found")
	}

	if err != nil {
		return "", fmt.Errorf("failed to get \"access_token\" ID: %w", err)
	}

	return accessToken, nil
}

// CreateuserToken creates a new user token with the given data and returns the created token.
func UpsertUserToken(ctx context.Context, tx *sqlx.Tx, data UserToken) (*UserToken, error) {
	q := `
        INSERT INTO user_access_tokens (id, access_token)
        VALUES ($1, $2)
		ON CONFLICT (id)
		DO UPDATE SET access_token = EXCLUDED.access_token
		RETURNING id, access_token;
    `

	var result UserToken
	err := tx.QueryRowContext(ctx, q,
		data.ID,
		data.AccessToken,
	).Scan(
		&result.ID,
		&result.AccessToken,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create \"remote_cluster_tokens\" entry: %w", err)
	}

	return &result, nil
}

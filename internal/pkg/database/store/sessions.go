package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/canonical/lxd/shared/api"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/types"
	"github.com/jmoiron/sqlx"
)

// Session represents the auth session associated with an identity.
type Session struct {
	ID              string    `db:"id"`          // Primary key
	IdentityID      int       `db:"identity_id"` // Foreign key to identities
	ExpiresAt       time.Time `db:"expires_at"`  // Session expiration timestamp
	CreatedAt       time.Time `db:"created_at"`  // Creation timestamp
	EncryptedTokens types.EncryptedTokenSet
}

// GetSessionIDByIdentityID returns the session ID by identity ID.
func GetSessionIDByIdentityID(ctx context.Context, tx *sqlx.Tx, identityID int) (string, error) {
	q := `
        SELECT id
        FROM sessions
        WHERE identity_id = $1
    `

	var id string
	err := tx.QueryRowContext(ctx, q, identityID).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return "", api.StatusErrorf(http.StatusNotFound, "session not found")
	}

	if err != nil {
		return "", fmt.Errorf("failed to get \"sessions\" ID: %w", err)
	}

	return id, nil
}

// SessionExistsForIdentity checks if a session exists for a given identity ID.
func SessionExistsForIdentity(ctx context.Context, tx *sqlx.Tx, identityID int) (bool, error) {
	_, err := GetSessionIDByIdentityID(ctx, tx, identityID)
	if err != nil {
		if api.StatusErrorCheck(err, http.StatusNotFound) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

// GetSessionByID returns a single session by session ID.
func GetSessionByID(ctx context.Context, tx *sqlx.Tx, sessionID string) (*Session, error) {
	q := `
		SELECT id, identity_id, expires_at, created_at, id_token_encrypted, access_token_encrypted, refresh_token_encrypted
		FROM sessions
		WHERE id = $1;
	`

	var result Session
	err := tx.QueryRowContext(ctx, q, sessionID).Scan(
		&result.ID,
		&result.IdentityID,
		&result.ExpiresAt,
		&result.CreatedAt,
		&result.EncryptedTokens.IDToken,
		&result.EncryptedTokens.AccessToken,
		&result.EncryptedTokens.RefreshToken,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, api.StatusErrorf(http.StatusNotFound, "session not found")
	}

	if err != nil {
		return nil, fmt.Errorf("failed to fetch from \"sessions\" table: %w", err)
	}

	return &result, nil
}

// CreateSession creates a new session.
func CreateSession(ctx context.Context, tx *sqlx.Tx, data Session) (*Session, error) {
	exists, err := SessionExistsForIdentity(ctx, tx, data.IdentityID)
	if err != nil {
		return nil, fmt.Errorf("failed to check for duplicates: %w", err)
	}

	if exists {
		return nil, api.StatusErrorf(http.StatusConflict, "this \"sessions\" entry already exists")
	}

	q := `
		INSERT INTO sessions (id, identity_id, expires_at, id_token_encrypted, access_token_encrypted, refresh_token_encrypted)
		VALUES ($1, $2, $3, $4, $5, $6)
        RETURNING id, identity_id, expires_at, created_at, id_token_encrypted, access_token_encrypted, refresh_token_encrypted;
    `

	var result Session
	err = tx.QueryRowContext(ctx, q,
		data.ID,
		data.IdentityID,
		data.ExpiresAt,
		data.EncryptedTokens.IDToken,
		data.EncryptedTokens.AccessToken,
		data.EncryptedTokens.RefreshToken,
	).Scan(
		&result.ID,
		&result.IdentityID,
		&result.ExpiresAt,
		&result.CreatedAt,
		&result.EncryptedTokens.IDToken,
		&result.EncryptedTokens.AccessToken,
		&result.EncryptedTokens.RefreshToken,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create \"sessions\" entry: %w", err)
	}

	return &result, nil
}

// UpdateSession updates an existing session by ID.
func UpdateSession(ctx context.Context, tx *sqlx.Tx, data Session) error {
	q := `
		UPDATE sessions
		SET expires_at = $1,
			id_token_encrypted = $2,
			access_token_encrypted = $3,
			refresh_token_encrypted = $4
		WHERE id = $5;
	`

	result, err := tx.ExecContext(ctx, q,
		data.ExpiresAt,
		data.EncryptedTokens.IDToken,
		data.EncryptedTokens.AccessToken,
		data.EncryptedTokens.RefreshToken,
		data.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update session: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get number of affected rows: %w", err)
	}

	if rows == 0 {
		return api.StatusErrorf(http.StatusNotFound, "session not found")
	}

	if rows > 1 {
		return fmt.Errorf("updated %d sessions instead of 1", rows)
	}

	return nil
}

// DeleteSession deletes a session by session ID.
func DeleteSession(ctx context.Context, tx *sqlx.Tx, sessionID string) error {
	q := `
        DELETE FROM sessions
        WHERE id = $1
    `

	result, err := tx.ExecContext(ctx, q, sessionID)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	n, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get number of affected rows: %w", err)
	}

	if n == 0 {
		return api.StatusErrorf(http.StatusNotFound, "no session found for session ID: %s", sessionID)
	} else if n > 1 {
		return fmt.Errorf("deleted %d sessions instead of 1", n)
	}

	return nil
}

// DeleteIdentitySessions deletes all sessions associated with a given identity.
func DeleteIdentitySessions(ctx context.Context, tx *sqlx.Tx, identityID int) error {
	q := `
		DELETE FROM sessions
		WHERE identity_id = $1
	`

	result, err := tx.ExecContext(ctx, q, identityID)
	if err != nil {
		return fmt.Errorf("failed to delete sessions for identity %d: %w", identityID, err)
	}

	n, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get number of affected rows: %w", err)
	}

	if n == 0 {
		return api.StatusErrorf(http.StatusNotFound, "no session found for identity ID: %d", identityID)
	}

	return nil
}

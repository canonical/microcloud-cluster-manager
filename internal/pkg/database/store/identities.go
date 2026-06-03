package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/canonical/lxd/shared/api"
	"github.com/jmoiron/sqlx"
)

// Identity represents a user identity mapped from an OIDC provider.
type Identity struct {
	ID          int       `db:"id"`           // Primary key
	Subject     string    `db:"subject"`      // Provider subject identifier
	Name        string    `db:"name"`         // Full name
	Email       string    `db:"email"`        // Email address
	DisplayName string    `db:"display_name"` // Display name
	CreatedAt   time.Time `db:"created_at"`   // Creation timestamp
	LastSeen    time.Time `db:"last_seen"`    // Last seen timestamp
	UpdatedAt   time.Time `db:"updated_at"`   // Update timestamp
}

// GetIdentityBySubject returns the ID of an identity by subject.
func GetIdentityBySubject(ctx context.Context, tx *sqlx.Tx, subject string) (*Identity, error) {
	q := `
        SELECT id
        FROM identities
        WHERE subject = $1
    `

	var id int
	err := tx.QueryRowContext(ctx, q, subject).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, api.StatusErrorf(http.StatusNotFound, "identity not found")
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get \"identities\" ID: %w", err)
	}

	return &Identity{ID: id}, nil
}

// IdentityExists checks if an identity with the given subject exists.
func IdentityExists(ctx context.Context, tx *sqlx.Tx, subject string) (bool, error) {
	_, err := GetIdentityBySubject(ctx, tx, subject)
	if err != nil {
		if api.StatusErrorCheck(err, http.StatusNotFound) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

// GetIdentity returns a single identity by subject.
func GetIdentity(ctx context.Context, tx *sqlx.Tx, subject string) (*Identity, error) {
	q := `
		SELECT id, subject, name, email, display_name, created_at, last_seen, updated_at
		FROM identities
		WHERE subject = $1;
	`

	var result Identity
	err := tx.QueryRowContext(ctx, q, subject).Scan(
		&result.ID,
		&result.Subject,
		&result.Name,
		&result.Email,
		&result.DisplayName,
		&result.CreatedAt,
		&result.LastSeen,
		&result.UpdatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, api.StatusErrorf(http.StatusNotFound, "identity not found")
	}

	if err != nil {
		return nil, fmt.Errorf("failed to fetch from \"identities\" table: %w", err)
	}

	return &result, nil
}

// CreateIdentity creates a new identity.
func CreateIdentity(ctx context.Context, tx *sqlx.Tx, data Identity) (*Identity, error) {
	exists, err := IdentityExists(ctx, tx, data.Subject)
	if err != nil {
		return nil, fmt.Errorf("failed to check for duplicates: %w", err)
	}

	if exists {
		return nil, api.StatusErrorf(http.StatusConflict, "this \"identities\" entry already exists")
	}

	q := `
        INSERT INTO identities (subject, name, email, display_name, last_seen)
        VALUES ($1, $2, $3, $4, $5)
        RETURNING id, subject, name, email, display_name, created_at, last_seen, updated_at;
    `

	var result Identity
	err = tx.QueryRowContext(ctx, q,
		data.Subject,
		data.Name,
		data.Email,
		data.DisplayName,
		data.LastSeen,
	).Scan(
		&result.ID,
		&result.Subject,
		&result.Name,
		&result.Email,
		&result.DisplayName,
		&result.CreatedAt,
		&result.LastSeen,
		&result.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create \"identities\" entry: %w", err)
	}

	return &result, nil
}

// UpdateIdentity updates an existing identity and returns the updated record.
func UpdateIdentity(ctx context.Context, tx *sqlx.Tx, data Identity) (*Identity, error) {
	q := `
		UPDATE identities
		SET name = $1,
			email = $2,
			display_name = $3,
			last_seen = $4,
			updated_at = NOW()
		WHERE id = $5
		RETURNING id, subject, name, email, display_name, created_at, last_seen, updated_at;
	`

	var result Identity
	err := tx.QueryRowContext(ctx, q,
		data.Name,
		data.Email,
		data.DisplayName,
		data.LastSeen,
		data.ID,
	).Scan(
		&result.ID,
		&result.Subject,
		&result.Name,
		&result.Email,
		&result.DisplayName,
		&result.CreatedAt,
		&result.LastSeen,
		&result.UpdatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, api.StatusErrorf(http.StatusNotFound, "identity not found")
	}

	if err != nil {
		return nil, fmt.Errorf("failed to update \"identities\" entry: %w", err)
	}

	return &result, nil
}

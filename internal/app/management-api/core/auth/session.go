package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/canonical/microcloud-cluster-manager/internal/pkg/database"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/database/store"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/types"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/zitadel/oidc/v3/pkg/oidc"
)

type SessionHandler struct {
	db *database.DB
}

// NewSessionHandler returns a new [SessionHandler].
func NewSessionHandler(db *database.DB) *SessionHandler {
	return &SessionHandler{db: db}
}

// StartSession creates a new session for the given OIDC tokens and returns the session ID and expiry time.
// It deletes any existing sessions for the associated identity to ensure only one active session per user.
// The tokens in encryptedTokens must be pre-encrypted by the caller before being passed in.
func (s *SessionHandler) StartSession(r *http.Request, claims *oidc.IDTokenClaims, encryptedTokens EncryptedTokenSet) (*uuid.UUID, *time.Time, error) {
	if s == nil || s.db == nil {
		return nil, nil, errors.New("Missing session handler")
	}

	if claims == nil {
		return nil, nil, errors.New("Missing OIDC claims")
	}

	if strings.TrimSpace(claims.Subject) == "" {
		return nil, nil, errors.New("Missing OIDC subject claim")
	}

	sessionID, err := uuid.NewV7()
	if err != nil {
		return nil, nil, fmt.Errorf("Failed creating new session UUID: %w", err)
	}

	sessionExpiry := time.Now().UTC().Add(OIDCSessionExpiry)

	err = s.db.Transaction(r.Context(), func(ctx context.Context, tx *sqlx.Tx) error {
		identity, err := getOrCreateIdentity(ctx, tx, claims)
		if err != nil {
			return err
		}

		err = store.DeleteIdentitySessions(ctx, tx, identity.ID)
		if err != nil && !store.IsNotFound(err) {
			return fmt.Errorf("Failed to delete old sessions by identity ID: %w", err)
		}

		_, err = store.CreateSession(ctx, tx, store.Session{
			ID:              sessionID.String(),
			IdentityID:      identity.ID,
			ExpiresAt:       sessionExpiry,
			EncryptedTokens: types.EncryptedTokenSet(encryptedTokens),
		})

		return err
	})

	if err != nil {
		return nil, nil, fmt.Errorf("Failed to create session: %w", err)
	}

	return &sessionID, &sessionExpiry, nil
}

// getOrCreateIdentity retrieves the identity for the given claims, creating it on first login if it does not yet exist.
func getOrCreateIdentity(ctx context.Context, tx *sqlx.Tx, claims *oidc.IDTokenClaims) (*store.Identity, error) {
	name := strings.TrimSpace(claims.Name)
	email := strings.TrimSpace(claims.Email)

	identity, err := store.GetIdentity(ctx, tx, claims.Subject)
	if err == nil {
		if name != "" {
			identity.Name = name
			identity.DisplayName = name
		}

		if email != "" {
			identity.Email = email
		}

		identity.LastSeen = time.Now().UTC()
		_, err = store.UpdateIdentity(ctx, tx, *identity)
		if err != nil {
			return nil, fmt.Errorf("Failed to update identity: %w", err)
		}
		return identity, nil
	}

	if !store.IsNotFound(err) {
		return nil, fmt.Errorf("Failed to get identity: %w", err)
	}

	if name == "" {
		name = claims.Subject
	}

	identity = &store.Identity{
		Subject:     claims.Subject,
		Name:        name,
		Email:       email,
		DisplayName: name,
		LastSeen:    time.Now().UTC(),
	}

	identity, err = store.CreateIdentity(ctx, tx, *identity)
	if err != nil {
		return nil, fmt.Errorf("Failed to create identity: %w", err)
	}

	return identity, nil
}

// GetSessionByID returns the session associated with the given session UUID.
func (s *SessionHandler) GetSessionByID(ctx context.Context, sessionID uuid.UUID) (*store.Session, error) {
	var session *store.Session

	err := s.db.Transaction(ctx, func(ctx context.Context, tx *sqlx.Tx) error {
		var err error
		session, err = store.GetSessionByID(ctx, tx, sessionID.String())
		if store.IsNotFound(err) {
			return err
		}
		if err != nil {
			return fmt.Errorf("Failed to get session by ID: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return session, nil
}

func (s *SessionHandler) DeleteSession(ctx context.Context, sessionID uuid.UUID) error {
	err := s.db.Transaction(ctx, func(ctx context.Context, tx *sqlx.Tx) error {
		err := store.DeleteSession(ctx, tx, sessionID.String())
		if err != nil {
			return fmt.Errorf("Failed to delete session: %w", err)
		}
		return nil
	})

	return err
}

// UpdateSessionTokens persists re-encrypted tokens for the given session and returns the session after it is updated.
func (s *SessionHandler) UpdateSessionTokens(ctx context.Context, session *store.Session, encryptedTokens EncryptedTokenSet) (*store.Session, error) {
	if s == nil || s.db == nil {
		return nil, errors.New("Missing session handler")
	}

	session.EncryptedTokens.IDToken = encryptedTokens.IDToken
	session.EncryptedTokens.AccessToken = encryptedTokens.AccessToken
	session.EncryptedTokens.RefreshToken = encryptedTokens.RefreshToken

	err := s.db.Transaction(ctx, func(ctx context.Context, tx *sqlx.Tx) error {
		return store.UpdateSession(ctx, tx, *session)
	})
	if err != nil {
		return nil, err
	}

	return session, nil
}

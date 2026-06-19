package auth

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

type sessionCredentialsContextKey string

// SessionCredentialsProviderKey is the key used to store the session credentials provider in the request context.
const SessionCredentialsProviderKey sessionCredentialsContextKey = "session-credentials-provider"

// SessionCredentialsProvider provides authenticated token operations for the current request session credentials.
type SessionCredentialsProvider struct {
	verifier             *Verifier
	sessionID            uuid.UUID
	userSecret           string
	encryptedAccessToken string
}

// GetAccessToken returns the decrypted token currently stored for the session credentials.
func (p *SessionCredentialsProvider) GetAccessToken() (string, error) {
	if p.encryptedAccessToken == "" {
		return "", fmt.Errorf("Missing encrypted access token")
	}

	accessToken, err := DecryptToken(p.encryptedAccessToken, p.userSecret)
	if err != nil {
		return "", err
	}

	return accessToken, nil
}

// RefreshSessionTokens refreshes OIDC tokens, persists them, and returns the refreshed token.
func (p *SessionCredentialsProvider) RefreshSessionTokens(ctx context.Context) (string, error) {
	session, err := p.verifier.sessionHandler.GetSessionByID(ctx, p.sessionID)
	if err != nil {
		return "", fmt.Errorf("Failed to get session: %w", err)
	}

	_, _, refreshToken, err := DecryptTokens(session.EncryptedTokens, p.userSecret)

	if err != nil {
		return "", err
	}

	_, encryptedTokens, refreshedAccessToken, err := p.verifier.refreshTokens(ctx, refreshToken, p.userSecret)
	if err != nil {
		return "", err
	}

	_, err = p.verifier.sessionHandler.UpdateSessionTokens(ctx, session, encryptedTokens)
	if err != nil {
		return "", fmt.Errorf("Failed to update session tokens: %w", err)
	}

	// Update cached encrypted access token.
	p.encryptedAccessToken = encryptedTokens.AccessToken

	return refreshedAccessToken, nil
}

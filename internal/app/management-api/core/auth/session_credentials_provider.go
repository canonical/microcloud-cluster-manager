package auth

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/zitadel/oidc/v3/pkg/client/rp"
)

type sessionCredentialsContextKey string

// SessionCredentialsProviderKey is the key used to store the session credentials provider in the request context.
const SessionCredentialsProviderKey sessionCredentialsContextKey = "session-credentials-provider"

// SessionCredentialsProvider provides authenticated token operations for the current request session credentials.
type SessionCredentialsProvider struct {
	verifier             *Verifier
	relyingParty         rp.RelyingParty
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

// GetUserSecret returns the user secret used for encrypting and decrypting tokens for the session credentials.
func (p *SessionCredentialsProvider) GetUserSecret() (string, error) {
	if p.userSecret == "" {
		return "", fmt.Errorf("Missing user secret")
	}

	return p.userSecret, nil
}

// RefreshSessionTokens refreshes OIDC tokens, persists them, and returns the refreshed token.
// If another request already refreshed the session's tokens, those are returned instead of redeeming the refresh
// token again.
func (p *SessionCredentialsProvider) RefreshSessionTokens(ctx context.Context) (string, error) {
	_, session, err := p.verifier.refreshSessionTokens(ctx, p.relyingParty, p.sessionID, p.userSecret, p.encryptedAccessToken)
	if err != nil {
		return "", err
	}

	refreshedAccessToken, err := DecryptToken(session.EncryptedTokens.AccessToken, p.userSecret)
	if err != nil {
		return "", err
	}

	// Update cached encrypted access token.
	p.encryptedAccessToken = session.EncryptedTokens.AccessToken

	return refreshedAccessToken, nil
}

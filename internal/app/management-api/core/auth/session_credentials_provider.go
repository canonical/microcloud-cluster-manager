package auth

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"
)

type sessionCredentialsContextKey string

// SessionCredentialsProviderKey is the key used to store the session credentials provider in the request context.
const SessionCredentialsProviderKey sessionCredentialsContextKey = "session-credentials-provider"

// SessionCredentialsProvider provides authenticated token operations for the current request session credentials.
type SessionCredentialsProvider struct {
	// mu guards encryptedAccessToken.
	mu sync.RWMutex
	// refreshMu serializes token refreshes.
	refreshMu            sync.Mutex
	verifier             *Verifier
	sessionID            uuid.UUID
	userSecret           string
	encryptedAccessToken string
}

// NewSessionCredentialsProvider returns a provider for the credentials of one session.
func NewSessionCredentialsProvider(verifier *Verifier, sessionID uuid.UUID, userSecret string, encryptedAccessToken string) *SessionCredentialsProvider {
	return &SessionCredentialsProvider{
		verifier:             verifier,
		sessionID:            sessionID,
		userSecret:           userSecret,
		encryptedAccessToken: encryptedAccessToken,
	}
}

// GetAccessToken returns the decrypted token currently stored for the session credentials.
func (p *SessionCredentialsProvider) GetAccessToken() (string, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
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

// RefreshSessionTokens refreshes OIDC tokens, persists them, and returns the refreshed access token.
// rejectedToken is the access token that was rejected. If the session already holds a different one, because this or
// another request refreshed it, that token is returned instead: refresh tokens may be single-use and must not be
// redeemed twice.
func (p *SessionCredentialsProvider) RefreshSessionTokens(ctx context.Context, rejectedToken string) (string, error) {
	p.refreshMu.Lock()
	defer p.refreshMu.Unlock()

	currentToken, err := p.GetAccessToken()
	if err == nil && currentToken != rejectedToken {
		return currentToken, nil
	}

	return p.refreshSessionTokens(ctx, rejectedToken)
}

func (p *SessionCredentialsProvider) refreshSessionTokens(ctx context.Context, rejectedToken string) (string, error) {
	session, err := p.verifier.sessionHandler.GetSessionByID(ctx, p.sessionID)
	if err != nil {
		return "", fmt.Errorf("Failed to get session: %w", err)
	}

	_, storedAccessToken, refreshToken, err := DecryptTokens(session.EncryptedTokens, p.userSecret)
	if err != nil {
		return "", err
	}

	// The cache of this provider only knows about refreshes made through it, the session knows about all of them.
	if storedAccessToken != rejectedToken {
		p.setEncryptedAccessToken(session.EncryptedTokens.AccessToken)
		return storedAccessToken, nil
	}

	_, encryptedTokens, refreshedAccessToken, err := p.verifier.refreshTokens(ctx, refreshToken, p.userSecret)
	if err != nil {
		return "", err
	}

	_, err = p.verifier.sessionHandler.UpdateSessionTokens(ctx, session, encryptedTokens)
	if err != nil {
		return "", fmt.Errorf("Failed to update session tokens: %w", err)
	}

	p.setEncryptedAccessToken(encryptedTokens.AccessToken)

	return refreshedAccessToken, nil
}

func (p *SessionCredentialsProvider) setEncryptedAccessToken(encryptedAccessToken string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.encryptedAccessToken = encryptedAccessToken
}

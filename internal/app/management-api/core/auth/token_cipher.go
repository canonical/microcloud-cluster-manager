package auth

import (
	"fmt"

	"github.com/canonical/microcloud-cluster-manager/internal/pkg/config"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/types"
)

type EncryptedTokenSet = types.EncryptedTokenSet

// EncryptTokens encrypts all three OIDC tokens using the given user secret.
func EncryptTokens(idToken, accessToken, refreshToken, userSecret string) (EncryptedTokenSet, error) {
	encryptedIDToken, err := config.EncryptUserValue(idToken, userSecret)
	if err != nil {
		return EncryptedTokenSet{}, fmt.Errorf("Failed to encrypt ID token: %w", err)
	}

	encryptedAccessToken, err := config.EncryptUserValue(accessToken, userSecret)
	if err != nil {
		return EncryptedTokenSet{}, fmt.Errorf("Failed to encrypt access token: %w", err)
	}

	encryptedRefreshToken, err := config.EncryptUserValue(refreshToken, userSecret)
	if err != nil {
		return EncryptedTokenSet{}, fmt.Errorf("Failed to encrypt refresh token: %w", err)
	}

	return EncryptedTokenSet{
		IDToken:      encryptedIDToken,
		AccessToken:  encryptedAccessToken,
		RefreshToken: encryptedRefreshToken,
	}, nil
}

// DecryptTokens decrypts all three encrypted tokens using the given user secret.
func DecryptTokens(ts EncryptedTokenSet, userSecret string) (idToken, accessToken, refreshToken string, err error) {
	idToken, err = config.DecryptUserValue(ts.IDToken, userSecret)
	if err != nil {
		return "", "", "", fmt.Errorf("Failed to decrypt ID token: %w", err)
	}

	accessToken, err = config.DecryptUserValue(ts.AccessToken, userSecret)
	if err != nil {
		return "", "", "", fmt.Errorf("Failed to decrypt access token: %w", err)
	}

	refreshToken, err = config.DecryptUserValue(ts.RefreshToken, userSecret)
	if err != nil {
		return "", "", "", fmt.Errorf("Failed to decrypt refresh token: %w", err)
	}

	return idToken, accessToken, refreshToken, nil
}

// DecryptToken decrypts any single encrypted token using the given user secret.
// This works for any token type (ID, access, or refresh) since they are all
// encrypted with the same user secret.
func DecryptToken(encryptedToken, userSecret string) (string, error) {
	token, err := config.DecryptUserValue(encryptedToken, userSecret)
	if err != nil {
		return "", fmt.Errorf("Failed to decrypt token: %w", err)
	}

	return token, nil
}

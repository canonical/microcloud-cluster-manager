package auth

import (
	"context"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/canonical/lxd/lxd/response"
	"github.com/canonical/lxd/lxd/util"
	"github.com/canonical/lxd/shared"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/config"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/database"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/database/store"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/logger"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/types"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/gorilla/securecookie"
	"github.com/zitadel/oidc/v3/pkg/client/rp"
	httphelper "github.com/zitadel/oidc/v3/pkg/http"
	"github.com/zitadel/oidc/v3/pkg/oidc"
	"golang.org/x/crypto/hkdf"
)

const (
	// cookieNameSessionToken is used to identify the session. It does not need to be encrypted.
	cookieNameSessionToken = "oidc_session"

	// cookieNameUserSecret is the identifier used to set and retrieve the user secret for encrypting tunnel data.
	cookieNameUserSecret = "user_secret"

	// SessionCookieExpiryBuffer extends browser cookie retention after session JWT expiry.
	// This helps ensure clients still send stale cookies so the server can respond with a deterministic re-authentication path.
	SessionCookieExpiryBuffer = time.Hour * 24 * 7

	// OIDCSessionExpiry is the duration for the session JWT expiry.
	OIDCSessionExpiry = time.Hour * 24 * 7
)

const (
	defaultConfigExpiryInterval = 5 * time.Minute
)

// Verifier holds all information needed to verify an access token offline.
type Verifier struct {
	relyingParty rp.RelyingParty

	clientID       string
	clientSecret   string
	issuer         string
	audience       string
	clusterCert    func() *shared.CertInfo
	httpClientFunc func() (*http.Client, error)

	// host is used for setting a valid callback URL when setting the relyingParty.
	// When creating the relyingParty, the OIDC library performs discovery (e.g. it calls the /well-known/oidc-configuration endpoint).
	// We don't want to perform this on every request, so we only do it when the request host changes.
	host string

	// configExpiry is the next time at which the relying party and access token verifier will be considered out of date
	// and will be refreshed. This refreshes the cookie encryption keys that the relying party uses.
	configExpiry         time.Time
	configExpiryInterval time.Duration

	sessionHandler *SessionHandler
}

// sessionTokenClaims represents claims for the signed session token.
type sessionTokenClaims struct {
	jwt.RegisteredClaims
}

// AuthError represents an authentication error. If an error of this type is returned, the caller should call
// WriteHeaders on the response so that the client has the necessary information to log in using the device flow.
type AuthError struct {
	Err error
}

// Error implements the error interface for AuthError.
func (e AuthError) Error() string {
	return fmt.Sprintf("Failed to authenticate: %s", e.Err.Error())
}

// Unwrap implements the xerrors.Wrapper interface for AuthError.
func (e AuthError) Unwrap() error {
	return e.Err
}

// StateToken is used to encode the state of the OIDC client in a URL which is used to prevent CSRF attacks (https://datatracker.ietf.org/doc/html/rfc6749#section-10.12).
// RedirectURL is the URL to which the client will be redirected after authentication.
// ID is a unique identifier for the state token and therefore the current login session.
type StateToken struct {
	RedirectURL string
	ID          string
}

// String encodes the StateToken as a base64 encoded string.
func (st StateToken) String() (string, error) {
	tokenData, err := json.Marshal(st)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(tokenData), nil
}

// DecodeStateToken decodes a base64 encoded string into a StateToken.
func DecodeStateToken(token string) (StateToken, error) {
	tokenData, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return StateToken{}, err
	}

	var stateToken StateToken
	err = json.Unmarshal(tokenData, &stateToken)
	if err != nil {
		return StateToken{}, err
	}

	return stateToken, nil
}

// Auth extracts and verifies the session token from the request cookie.
// If the session token is expired or invalid, it returns an error which can be used to trigger the OIDC login flow on the client.
func (o *Verifier) Auth(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	err := o.ensureConfig(ctx, r)
	if err != nil {
		return AuthError{Err: fmt.Errorf("Failed to ensure config: %w", err)}
	}

	sessionCookie, err := r.Cookie(cookieNameSessionToken)
	if err != nil {
		return AuthError{Err: fmt.Errorf("Failed to get session cookie: %w", err)}
	}

	claims, sessionID, userSecret, encryptedAccessToken, err := o.authenticate(ctx, w, r, sessionCookie.Value)
	if err != nil {
		return AuthError{Err: err}
	}

	groups := extractIdpGroups(claims)
	setUserInfoInRequest(claims.Email, claims.Name, groups, r)

	provider := &SessionCredentialsProvider{
		verifier:             o,
		sessionID:            sessionID,
		userSecret:           userSecret,
		encryptedAccessToken: encryptedAccessToken,
	}

	providerCtx := context.WithValue(r.Context(), SessionCredentialsProviderKey, provider)
	*r = *r.WithContext(providerCtx)

	return nil
}

// authenticate resolves a session token to verified OIDC claims, handling the expired-session renewal path.
func (o *Verifier) authenticate(ctx context.Context, w http.ResponseWriter, r *http.Request, sessionToken string) (*oidc.IDTokenClaims, uuid.UUID, string, string, error) {
	sessionID, err := o.verifySessionAndGetID(sessionToken)
	if err != nil {
		if !errors.Is(err, jwt.ErrTokenExpired) {
			return nil, uuid.Nil, "", "", fmt.Errorf("Failed to verify session token: %w", err)
		}

		claims, newSessionID, userSecret, encryptedAccessToken, err := o.handleExpiredSession(ctx, r, w, sessionID)
		if err != nil {
			return nil, uuid.Nil, "", "", fmt.Errorf("Failed to handle expired session: %w", err)
		}

		return claims, newSessionID, userSecret, encryptedAccessToken, nil
	}

	userSecret, err := o.getUserSecretFromCookie(r, sessionID)
	if err != nil {
		return nil, uuid.Nil, "", "", fmt.Errorf("Failed to get user secret: %w", err)
	}

	session, err := o.sessionHandler.GetSessionByID(r.Context(), sessionID)
	if store.IsNotFound(err) {
		o.deleteCookies(w)
	}

	if err != nil {
		return nil, uuid.Nil, "", "", fmt.Errorf("Failed to get session: %w", err)
	}

	// Decrypt tokens stored in the database using the per-user secret from the cookie.
	idToken, _, refreshToken, err := DecryptTokens(session.EncryptedTokens, userSecret)

	if err != nil {
		return nil, uuid.Nil, "", "", err
	}

	claims, err := rp.VerifyIDToken[*oidc.IDTokenClaims](ctx, idToken, o.relyingParty.IDTokenVerifier())
	if err != nil {
		claims, encryptedTokens, _, err := o.refreshTokens(ctx, refreshToken, userSecret)
		if err != nil {
			return nil, uuid.Nil, "", "", err
		}

		_, err = o.sessionHandler.UpdateSessionTokens(ctx, session, encryptedTokens)
		if err != nil {
			return nil, uuid.Nil, "", "", fmt.Errorf("Failed to update session tokens: %w", err)
		}

		return claims, sessionID, userSecret, encryptedTokens.AccessToken, nil
	}

	return claims, sessionID, userSecret, session.EncryptedTokens.AccessToken, nil
}

// verifySessionAndGetID verifies a signed session token and returns the embedded session ID.
func (o *Verifier) verifySessionAndGetID(sessionToken string) (uuid.UUID, error) {
	claims := &sessionTokenClaims{}
	parser := jwt.NewParser(jwt.WithValidMethods([]string{jwt.SigningMethodHS512.Alg()}))
	var sessionID uuid.UUID
	var err error

	_, err = parser.ParseWithClaims(sessionToken, claims, func(token *jwt.Token) (any, error) {
		typedClaims, ok := token.Claims.(*sessionTokenClaims)
		if !ok {
			return nil, errors.New("Malformed session JWT claims")
		}

		sessionIDClaim := typedClaims.Subject
		if sessionIDClaim == "" {
			return nil, errors.New("Missing session ID in JWT payload")
		}

		sessionID, err = uuid.Parse(sessionIDClaim)
		if err != nil {
			return nil, fmt.Errorf("Invalid session ID in JWT payload: %w", err)
		}

		key, err := o.getSessionSigningKey(sessionID)
		if err != nil {
			return nil, err
		}

		return key, nil
	})

	if err != nil {
		return sessionID, fmt.Errorf("Invalid session JWT: %w", err)
	}

	return sessionID, nil
}

// handleExpiredSession attempts to start a new session based on details saved in an existing session.
// It returns refreshed OIDC claims and session metadata so the caller can complete request context setup.
func (o *Verifier) handleExpiredSession(ctx context.Context, r *http.Request, w http.ResponseWriter, sessionID uuid.UUID) (*oidc.IDTokenClaims, uuid.UUID, string, string, error) {
	defer func() {
		err := o.sessionHandler.DeleteSession(ctx, sessionID)
		if err != nil && !store.IsNotFound(err) {
			logger.Log.Infof("Failed deleting session after expired session handling: %v", err)
		}
	}()

	session, err := o.sessionHandler.GetSessionByID(ctx, sessionID)
	if err != nil {
		return nil, uuid.Nil, "", "", fmt.Errorf("Failed to get session for expired session handling: %w", err)
	}

	userSecret, err := o.getUserSecretFromCookie(r, sessionID)
	if err != nil {
		return nil, uuid.Nil, "", "", fmt.Errorf("Failed to get user secret: %w", err)
	}

	idToken, _, refreshToken, err := DecryptTokens(session.EncryptedTokens, userSecret)

	if err != nil {
		return nil, uuid.Nil, "", "", fmt.Errorf("Failed to decrypt tokens for expired session handling: %w", err)
	}

	claims, err := rp.VerifyIDToken[*oidc.IDTokenClaims](ctx, idToken, o.relyingParty.IDTokenVerifier())
	if err != nil {
		var encryptedTokens EncryptedTokenSet
		claims, encryptedTokens, _, err = o.refreshTokens(ctx, refreshToken, userSecret)
		if err != nil {
			return nil, uuid.Nil, "", "", fmt.Errorf("Failed to refresh tokens: %w", err)
		}

		session, err = o.sessionHandler.UpdateSessionTokens(ctx, session, encryptedTokens)
		if err != nil {
			return nil, uuid.Nil, "", "", fmt.Errorf("Failed to update session tokens: %w", err)
		}
	}

	newSessionID, expiry, err := o.sessionHandler.StartSession(r, claims, session.EncryptedTokens)
	if err != nil {
		return nil, uuid.Nil, "", "", fmt.Errorf("Failed to start session: %w", err)
	}

	err = o.setCookies(w, *newSessionID, userSecret, *expiry)
	if err != nil {
		return nil, uuid.Nil, "", "", err
	}

	return claims, *newSessionID, userSecret, session.EncryptedTokens.AccessToken, nil
}

// deleteCookies expires all session-related cookies in the response, effectively deleting them on the client.
func (o *Verifier) deleteCookies(w http.ResponseWriter) {
	for _, name := range []string{cookieNameSessionToken, cookieNameUserSecret} {
		http.SetCookie(w, &http.Cookie{
			Name:     name,
			Path:     "/",
			Secure:   true,
			HttpOnly: true,
			SameSite: http.SameSiteStrictMode,
			Value:    "",
			Expires:  time.Unix(0, 0),
		})
	}
}

// refreshTokens fetches a new token set using the refresh token, verifies the refreshed ID token,
// and re-encrypts all tokens with the user secret. It should only be called after rp.VerifyIDToken
// has already failed. Persistence of the returned EncryptedTokenSet is the caller's responsibility.
func (o *Verifier) refreshTokens(ctx context.Context, refreshToken, userSecret string) (*oidc.IDTokenClaims, EncryptedTokenSet, string, error) {
	if refreshToken == "" {
		return nil, EncryptedTokenSet{}, "", errors.New("Missing refresh token")
	}

	tokens, err := rp.RefreshTokens[*oidc.IDTokenClaims](ctx, o.relyingParty, refreshToken, "", "")
	if err != nil {
		return nil, EncryptedTokenSet{}, "", fmt.Errorf("Failed to refresh ID tokens: %w", err)
	}

	refreshedIDToken, refreshedAccessToken, refreshedRefreshToken, err := extractTokensFromRefreshResponse(tokens, refreshToken)
	if err != nil {
		return nil, EncryptedTokenSet{}, "", fmt.Errorf("Failed to extract refreshed tokens from OIDC response: %w", err)
	}

	claims, err := rp.VerifyIDToken[*oidc.IDTokenClaims](ctx, refreshedIDToken, o.relyingParty.IDTokenVerifier())
	if err != nil {
		return nil, EncryptedTokenSet{}, "", fmt.Errorf("Failed to verify refreshed ID token: %w", err)
	}

	encryptedTokens, err := EncryptTokens(refreshedIDToken, refreshedAccessToken, refreshedRefreshToken, userSecret)
	if err != nil {
		return nil, EncryptedTokenSet{}, "", fmt.Errorf("Failed to encrypt refreshed tokens: %w", err)
	}

	return claims, encryptedTokens, refreshedAccessToken, nil
}

// extractTokensFromRefreshResponse extracts token strings from the OIDC refresh response.
func extractTokensFromRefreshResponse(tokens *oidc.Tokens[*oidc.IDTokenClaims], oldRefreshToken string) (idToken, accessToken, refreshToken string, err error) {
	idTokenAny := tokens.Extra("id_token")
	if idTokenAny == nil {
		return "", "", "", errors.New("ID tokens missing from OIDC refresh response")
	}

	refreshedIDToken, ok := idTokenAny.(string)
	if !ok {
		return "", "", "", errors.New("Malformed ID tokens in OIDC refresh response")
	}

	refreshedAccessToken := tokens.AccessToken
	if refreshedAccessToken == "" {
		return "", "", "", errors.New("Access token missing from OIDC refresh response")
	}

	refreshedRefreshToken := tokens.RefreshToken
	if refreshedRefreshToken == "" {
		// Some providers don't return a new refresh token on refresh; keep the previous value.
		refreshedRefreshToken = oldRefreshToken
	}

	return refreshedIDToken, refreshedAccessToken, refreshedRefreshToken, nil
}

// extractIdpGroups extracts the user's groups from the OIDC claims.
func extractIdpGroups(claims *oidc.IDTokenClaims) []string {
	raw, ok := claims.Claims["mcm-idp-groups"]
	if !ok {
		logger.Log.Info(`AUTHN "mcm-idp-groups" OIDC groups claim missing`)
		return nil
	}

	rawSlice, ok := raw.([]any)
	if !ok {
		logger.Log.Info("AUTHN OIDC groups claim malformed")
		return nil
	}
	groups := make([]string, 0, len(rawSlice))
	for _, v := range rawSlice {
		s, ok := v.(string)
		if !ok {
			logger.Log.Info("AUTHN OIDC groups claim malformed")
			return nil
		}
		groups = append(groups, s)
	}

	return groups
}

// Login is a http.Handler than initiates the login flow for the UI.
func (o *Verifier) Login(w http.ResponseWriter, r *http.Request, stateTokenStr string) {
	err := o.ensureConfig(r.Context(), r)
	if err != nil {
		logger.Log.Info("AUTHN invalid OIDC configuration")
		err := response.ErrorResponse(http.StatusInternalServerError, fmt.Errorf("Login failed: %w", err).Error()).Render(w, r)
		if err != nil {
			logger.Log.Errorw("Failed rendering internal server error response due to invalid OIDC configuration: %w", err)
		}
		return
	}

	logger.Log.Info("AUTHN initiating OIDC login flow")
	handler := rp.AuthURLHandler(func() string { return stateTokenStr }, o.relyingParty, rp.WithURLParam("audience", o.audience))
	handler(w, r)
}

// Logout always deletes the session cookie. If the caller is logged in with a valid session cookie, then that session is
// deleted from the database.
func (o *Verifier) Logout(w http.ResponseWriter, r *http.Request) {
	defer func() {
		o.deleteCookies(w)
		http.Redirect(w, r, "/ui/login/", http.StatusFound)
	}()

	sessionCookie, err := r.Cookie(cookieNameSessionToken)

	if err != nil {
		// Not logged in.
		return
	}

	sessionID, err := o.verifySessionAndGetID(sessionCookie.Value)
	if err != nil && !errors.Is(err, jwt.ErrTokenExpired) {
		// Not logged in.
		return
	}

	err = o.sessionHandler.DeleteSession(r.Context(), sessionID)

	if err != nil {
		logger.Log.Errorw("Failed to delete session information", "error", err)
		return
	}
}

// Callback is a http.HandlerFunc which implements the code exchange required on the /oidc/callback endpoint.
func (o *Verifier) Callback(w http.ResponseWriter, r *http.Request, redirectURL string) {
	err := o.ensureConfig(r.Context(), r)
	if err != nil {
		logger.Log.Info("AUTHN invalid OIDC configuration")
		err := response.ErrorResponse(http.StatusInternalServerError, fmt.Errorf("OIDC callback failed: %w", err).Error()).Render(w, r)
		if err != nil {
			logger.Log.Errorw("Failed rendering internal server error response due to invalid OIDC configuration: %w", err)
		}
		return
	}

	handler := rp.CodeExchangeHandler(func(w http.ResponseWriter, r *http.Request, tokens *oidc.Tokens[*oidc.IDTokenClaims], state string, relyingParty rp.RelyingParty) {
		verifiedClaims, err := rp.VerifyIDToken[*oidc.IDTokenClaims](r.Context(), tokens.IDToken, o.relyingParty.IDTokenVerifier())
		if err != nil {
			logger.Log.Info("AUTHN failed to verify OIDC ID token")
			err = response.ErrorResponse(http.StatusInternalServerError, fmt.Errorf("Failed to verify ID token: %w", err).Error()).Render(w, r)
			if err != nil {
				logger.Log.Errorw("Failed rendering internal server error response due to failed ID token verification: %w", err)
			}
			return
		}

		userSecret, err := config.CreateUserSecret()
		if err != nil {
			logger.Log.Info("AUTHN failed to generate user secret")
			err = response.ErrorResponse(http.StatusInternalServerError, fmt.Errorf("Failed to generate user secret: %w", err).Error()).Render(w, r)
			if err != nil {
				logger.Log.Errorw("Failed rendering internal server error response due to failed user secret generation: %w", err)
			}
			return
		}
		encryptedTokens, err := EncryptTokens(tokens.IDToken, tokens.AccessToken, tokens.RefreshToken, userSecret)
		if err != nil {
			logger.Log.Infof("AUTHN failed to encrypt tokens %v", err)
			err = response.ErrorResponse(http.StatusInternalServerError, fmt.Errorf("Failed to encrypt tokens: %w", err).Error()).Render(w, r)
			if err != nil {
				logger.Log.Errorw("Failed rendering internal server error response due to failed token encryption: %w", err)
			}
			return
		}

		sessionID, expiry, err := o.sessionHandler.StartSession(r, verifiedClaims, encryptedTokens)
		if err != nil {
			logger.Log.Infof("AUTHN failed to start session, %v", err)
			err = response.ErrorResponse(http.StatusInternalServerError, fmt.Errorf("Failed to start session: %w", err).Error()).Render(w, r)
			if err != nil {
				logger.Log.Errorw("Failed rendering internal server error response due to failed session start: %w", err)
			}
			return
		}

		err = o.setCookies(w, *sessionID, userSecret, *expiry)

		if err != nil {
			logger.Log.Infof("AUTHN failed to set cookies, %v", err)
			err = response.ErrorResponse(http.StatusInternalServerError, err.Error()).Render(w, r)
			if err != nil {
				logger.Log.Errorw("Failed rendering internal server error response due to failed cookie setting: %w", err)
			}
			return
		}

		// Send to the UI.
		// NOTE: Once the UI does the redirection on its own, we may be able to use the referer here instead.
		logger.Log.Info("AUTHN login successful, redirecting user")
		http.Redirect(w, r, redirectURL, http.StatusMovedPermanently)
	}, o.relyingParty)

	handler(w, r)
}

// getSignedSession creates a signed JWT with the value session ID.
func (o *Verifier) getSignedSession(sessionID uuid.UUID, expiry time.Time) (string, error) {
	key, err := o.getSessionSigningKey(sessionID)
	if err != nil {
		return "", err
	}

	claims := &sessionTokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   sessionID.String(),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(expiry),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)

	signedToken, err := token.SignedString(key)
	if err != nil {
		return "", fmt.Errorf("Failed to sign session JWT: %w", err)
	}

	return signedToken, nil
}

// getSessionSigningKey derives a 64-byte signing key for a session token using the management API private key and
// the session ID as HKDF salt.
func (o *Verifier) getSessionSigningKey(sessionID uuid.UUID) ([]byte, error) {
	salt, err := sessionID.MarshalBinary()
	if err != nil {
		return nil, fmt.Errorf("Failed to marshal session ID as binary: %w", err)
	}

	return deriveHKDFKey(o.clusterCert().PrivateKey(), salt, "SESSION_ID_TOKEN_INTEGRITY", 64)
}

// ExpireConfig sets the expiry time of the current configuration to zero. This forces the verifier to reconfigure the
// relying party the next time a user authenticates.
func (o *Verifier) ExpireConfig() {
	o.configExpiry = time.Now()
}

// ensureConfig ensures that the relyingParty and accessTokenVerifier fields of the Verifier are non-nil. Additionally,
// if the given host is different from the Verifier host we reset the relyingParty to ensure the callback URL is set
// correctly.
func (o *Verifier) ensureConfig(ctx context.Context, r *http.Request) error {
	if o.relyingParty == nil || r.Host != o.host || time.Now().After(o.configExpiry) {
		err := o.setRelyingParty(ctx, r)
		if err != nil {
			return err
		}

		o.host = r.Host
		o.configExpiry = time.Now().Add(o.configExpiryInterval)
	}

	return nil
}

// setRelyingParty sets the relyingParty on the Verifier. The request argument is used to set a valid callback URL.
func (o *Verifier) setRelyingParty(ctx context.Context, r *http.Request) error {
	// The relying party sets cookies for the following values:
	// - "state": Used to prevent CSRF attacks (https://datatracker.ietf.org/doc/html/rfc6749#section-10.12).
	// - "pkce": Used to prevent authorization code interception attacks (https://datatracker.ietf.org/doc/html/rfc7636).
	// Both should be stored securely. However, these cookies do not need to be decrypted by other cluster members, so
	// it is ok to use the secure key generation that is built in to the securecookie library. This also reduces the
	// exposure of our private key.

	// The hash key should be 64 bytes (https://github.com/gorilla/securecookie).
	cookieHashKey := securecookie.GenerateRandomKey(64)
	if cookieHashKey == nil {
		return errors.New("Failed to generate a secure cookie hash key")
	}

	// The block key should 32 bytes for AES-256 encryption.
	cookieBlockKey := securecookie.GenerateRandomKey(32)
	if cookieBlockKey == nil {
		return errors.New("Failed to generate a secure cookie hash key")
	}

	httpClient, err := o.httpClientFunc()
	if err != nil {
		return fmt.Errorf("Failed to get a HTTP client: %w", err)
	}

	cookieHandler := httphelper.NewCookieHandler(cookieHashKey, cookieBlockKey)
	options := []rp.Option{
		rp.WithCookieHandler(cookieHandler),
		rp.WithVerifierOpts(rp.WithIssuedAtOffset(5 * time.Second)),
		rp.WithPKCE(cookieHandler),
		rp.WithHTTPClient(httpClient),
	}

	oidcScopes := []string{oidc.ScopeOpenID, oidc.ScopeOfflineAccess, oidc.ScopeEmail, oidc.ScopeProfile}

	callbackURL := getCallbackURL(r.Host)

	relyingParty, err := rp.NewRelyingPartyOIDC(ctx, o.issuer, o.clientID, o.clientSecret, callbackURL, oidcScopes, options...)
	if err != nil {
		return fmt.Errorf("Failed to get OIDC relying party: %w", err)
	}

	o.relyingParty = relyingParty
	return nil
}

// getUserSecretFromCookie retrieves and decrypts the user secret from the request cookies.
func (o *Verifier) getUserSecretFromCookie(r *http.Request, sessionID uuid.UUID) (string, error) {
	cookie, err := r.Cookie(cookieNameUserSecret)
	if err != nil {
		return "", fmt.Errorf("Failed to get user secret cookie: %w", err)
	}

	sc, err := o.secureCookieFromSession(sessionID)
	if err != nil {
		return "", fmt.Errorf("Failed to create secure cookie for user secret: %w", err)
	}

	var userSecret string
	err = sc.Decode(cookieNameUserSecret, cookie.Value, &userSecret)
	if err != nil {
		return "", fmt.Errorf("Failed to decode user secret cookie: %w", err)
	}

	return userSecret, nil
}

// setCookies sets the session token and user secret cookies in the response.
func (o *Verifier) setCookies(w http.ResponseWriter, sessionID uuid.UUID, userSecret string, expiry time.Time) error {
	cookieExpiry := expiry.Add(SessionCookieExpiryBuffer)

	sc, err := o.secureCookieFromSession(sessionID)
	if err != nil {
		return fmt.Errorf("Failed to create secure cookie for user secret: %w", err)
	}

	maxAge := int(time.Until(cookieExpiry).Seconds())
	if maxAge < 0 {
		maxAge = 0
	}

	// Keep securecookie timestamp validation aligned with HTTP cookie expiry.
	sc.MaxAge(maxAge)

	encodedUserSecret, err := sc.Encode(cookieNameUserSecret, userSecret)
	if err != nil {
		return fmt.Errorf("Failed to encode user secret cookie: %w", err)
	}

	sessionToken, err := o.getSignedSession(sessionID, expiry)
	if err != nil {
		return fmt.Errorf("Failed to create session token: %w", err)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     cookieNameUserSecret,
		Value:    encodedUserSecret,
		Path:     "/",
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Expires:  cookieExpiry,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     cookieNameSessionToken,
		Path:     "/",
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Value:    sessionToken,
		Expires:  cookieExpiry,
	})

	return nil
}

// secureCookieFromSession returns a *securecookie.SecureCookie that is secure, unique to each client, and
// reproducible on all cluster members. Keys are derived from the cluster private key and the session ID using
// HKDF-SHA-512 (https://datatracker.ietf.org/doc/html/rfc5869), scoped to distinct purposes via the info parameter.
// Warning: Changes to this function might cause all existing OIDC users to be logged out of LXD (but not logged out of
// the IdP).
func (o *Verifier) secureCookieFromSession(sessionID uuid.UUID) (*securecookie.SecureCookie, error) {
	salt, err := sessionID.MarshalBinary()
	if err != nil {
		return nil, fmt.Errorf("Failed to marshal session ID as binary: %w", err)
	}

	clusterPrivateKey := o.clusterCert().PrivateKey()

	// 64-byte HMAC key for integrity verification (https://github.com/gorilla/securecookie).
	cookieHashKey, err := deriveHKDFKey(clusterPrivateKey, salt, "INTEGRITY", 64)
	if err != nil {
		return nil, fmt.Errorf("Failed creating secure cookie hash key: %w", err)
	}

	// 32-byte key for AES-256 encryption.
	cookieBlockKey, err := deriveHKDFKey(clusterPrivateKey, salt, "ENCRYPTION", 32)
	if err != nil {
		return nil, fmt.Errorf("Failed creating secure cookie block key: %w", err)
	}

	return securecookie.New(cookieHashKey, cookieBlockKey), nil
}

// Host returns the host of the Verifier.
func (o *Verifier) Host() string {
	return o.host
}

// NewVerifier returns a Verifier.
func NewVerifier(issuer string, clientID string, clientSecret string, audience string, cert *shared.CertInfo, db *database.DB) (*Verifier, error) {
	// Setup a http client for communicating with the OIDC provider.
	httpClientFunc := func() (*http.Client, error) {
		client, err := util.HTTPClient("", http.ProxyFromEnvironment)
		if err != nil {
			return nil, err
		}

		// NOTE: the http client we use to make requests to the OIDC provider must have the CA cert we create in the k8s cluster
		existingTransport, ok := client.Transport.(*http.Transport)
		if !ok {
			return nil, fmt.Errorf("unexpected transport type: %T", client.Transport)
		}

		newTransport := existingTransport.Clone()
		clientTLSConfig, err := shared.GetTLSConfig(nil)
		if err != nil {
			return nil, err
		}

		newTransport.TLSClientConfig = clientTLSConfig
		client.Transport = newTransport

		return client, nil
	}

	certFunc := func() *shared.CertInfo {
		return cert
	}

	verifier := &Verifier{
		issuer:               issuer,
		clientID:             clientID,
		clientSecret:         clientSecret,
		audience:             audience,
		clusterCert:          certFunc,
		configExpiryInterval: defaultConfigExpiryInterval,
		httpClientFunc:       httpClientFunc,
		sessionHandler:       NewSessionHandler(db),
	}

	return verifier, nil
}

// deriveHKDFKey derives a cryptographic key of the specified length from ikm (input key material) using HKDF-SHA-512.
// salt differentiates derived keys per session; info scopes the key to a specific purpose, preventing cross-context reuse.
// See https://datatracker.ietf.org/doc/html/rfc5869 for the HKDF specification.
func deriveHKDFKey(ikm, salt []byte, info string, length int) ([]byte, error) {
	prk := hkdf.Extract(sha512.New, ikm, salt)
	kdf := hkdf.Expand(sha512.New, prk, []byte(info))
	key := make([]byte, length)
	_, err := io.ReadFull(kdf, key)
	if err != nil {
		return nil, fmt.Errorf("Failed to derive HKDF key for %q: %w", info, err)
	}

	return key, nil
}

func getCallbackURL(host string) string {
	return fmt.Sprintf("https://%s/oidc/callback", host)
}

func setUserInfoInRequest(email, name string, groups []string, r *http.Request) {
	userInfo := &types.UserInfo{
		Email:  email,
		Name:   name,
		Groups: groups,
	}

	userInfoCtx := context.WithValue(r.Context(), types.UserInfoKey, userInfo)
	*r = *r.WithContext(userInfoCtx)
}

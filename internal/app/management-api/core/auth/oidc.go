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
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/logger"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/types"
	"github.com/google/uuid"
	"github.com/gorilla/securecookie"
	"github.com/zitadel/oidc/v3/pkg/client/rp"
	httphelper "github.com/zitadel/oidc/v3/pkg/http"
	"github.com/zitadel/oidc/v3/pkg/oidc"
	"golang.org/x/crypto/hkdf"
)

const (
	// cookieNameIDToken is the identifier used to set and retrieve the identity token.
	cookieNameIDToken = "oidc_identity"

	// cookieNameRefreshToken is the identifier used to set and retrieve the refresh token.
	cookieNameRefreshToken = "oidc_refresh"

	// cookieNameSessionID is used to identify the session. It does not need to be encrypted.
	cookieNameSessionID = "session_id"
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
}

// AuthenticationResult represents an authenticated OIDC client.
type AuthenticationResult struct {
	Subject string
	Email   string
	Name    string
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

// Auth extracts OIDC tokens from the request, verifies them, and returns the subject.
func (o *Verifier) Auth(ctx context.Context, w http.ResponseWriter, r *http.Request) (bool, error) {
	err := o.ensureConfig(ctx, r)
	if err != nil {
		return false, fmt.Errorf("Authorization failed: %w", err)
	}

	_, idToken, refreshToken, err := o.getCookies(r)
	if err != nil {
		// Cookies are present but we failed to decrypt them. They may have been tampered with, so delete them to force
		// the user to log in again.
		err := o.setCookies(w, nil, uuid.UUID{}, "", "", true)
		if err != nil {
			return false, fmt.Errorf("Failed to delete invalid login cookies: %w", err)
		}
		return false, fmt.Errorf("Failed to retrieve login information: %w", err)
	}

	if idToken == "" && refreshToken == "" {
		return false, fmt.Errorf("No ID or refresh tokens present in request")
	}

	// When authenticating via the UI, we expect that there will be ID and refresh tokens present in the request cookies.
	result, err := o.authenticateIDToken(ctx, w, idToken, refreshToken)
	if err != nil {
		return false, err
	}

	setUserInfoInRequest(result, r)

	return true, nil
}

// authenticateIDToken verifies the identity token and returns the ID token subject. If no identity token is given (or
// verification fails) it will attempt to refresh the ID token.
func (o *Verifier) authenticateIDToken(ctx context.Context, w http.ResponseWriter, idToken string, refreshToken string) (result *AuthenticationResult, e error) {
	var claims *oidc.IDTokenClaims
	var err error
	if idToken != "" {
		// Try to verify the ID token.
		claims, err = rp.VerifyIDToken[*oidc.IDTokenClaims](ctx, idToken, o.relyingParty.IDTokenVerifier())
		if err == nil {
			return &AuthenticationResult{
				Subject: claims.Subject,
				Email:   claims.Email,
				Name:    claims.Name,
			}, nil
		}
	}

	// If ID token verification failed (or it wasn't provided, try refreshing the token).
	tokens, err := rp.RefreshTokens[*oidc.IDTokenClaims](ctx, o.relyingParty, refreshToken, "", "")
	if err != nil {
		return nil, AuthError{Err: fmt.Errorf("Failed to refresh ID tokens: %w", err)}
	}

	idTokenAny := tokens.Extra("id_token")
	if idTokenAny == nil {
		return nil, AuthError{Err: errors.New("ID tokens missing from OIDC refresh response")}
	}

	idToken, ok := idTokenAny.(string)
	if !ok {
		return nil, AuthError{Err: errors.New("Malformed ID tokens in OIDC refresh response")}
	}

	// Verify the refreshed ID token.
	claims, err = rp.VerifyIDToken[*oidc.IDTokenClaims](ctx, idToken, o.relyingParty.IDTokenVerifier())
	if err != nil {
		return nil, AuthError{Err: fmt.Errorf("Failed to verify refreshed ID token: %w", err)}
	}

	// Updated the cookies
	err = o.WriteTokenToCookies(w, idToken, tokens.RefreshToken)
	if err != nil {
		return nil, AuthError{Err: fmt.Errorf("Failed to update login cookies: %w", err)}
	}

	return &AuthenticationResult{
		Subject: claims.Subject,
		Email:   claims.Email,
		Name:    claims.Name,
	}, nil
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

// Logout deletes the ID and refresh token cookies and redirects the user to the login page.
func (o *Verifier) Logout(w http.ResponseWriter, r *http.Request) {
	err := o.setCookies(w, nil, uuid.UUID{}, "", "", true)
	if err != nil {
		logger.Log.Errorw("Failed to delete login information: %w", err)
		err = response.ErrorResponse(http.StatusInternalServerError, fmt.Errorf("Failed to delete login information: %w", err).Error()).Render(w, r)
		if err != nil {
			logger.Log.Errorw("Failed rendering internal server error response due to failed deletion of login information: %w", err)
		}
		return
	}

	http.Redirect(w, r, "/ui/login", http.StatusFound)
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

	handler := rp.CodeExchangeHandler(func(w http.ResponseWriter, r *http.Request, tokens *oidc.Tokens[*oidc.IDTokenClaims], state string, rp rp.RelyingParty) {
		err := o.WriteTokenToCookies(w, tokens.IDToken, tokens.RefreshToken)

		if err != nil {
			logger.Log.Info("AUTHN failed to write OIDC tokens to cookies")
			err = response.ErrorResponse(http.StatusInternalServerError, err.Error()).Render(w, r)
			if err != nil {
				logger.Log.Errorw("Failed rendering internal server error response due to failed writing of OIDC tokens to cookies: %w", err)
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

// WriteTokenToCookies writes the ID and refresh tokens to the response cookies based on a new session ID.
// The cookies are encrypted using the cluster private key.
func (o *Verifier) WriteTokenToCookies(w http.ResponseWriter, idToken string, refreshToken string) error {
	sessionID := uuid.New()
	secureCookie, err := o.secureCookieFromSession(sessionID)
	if err != nil {
		return fmt.Errorf("Failed to create new session: %w", err)
	}

	err = o.setCookies(w, secureCookie, sessionID, idToken, refreshToken, false)
	if err != nil {
		return fmt.Errorf("Failed to update login cookies: %w", err)
	}

	return nil
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

// getCookies gets the sessionID, identity and refresh tokens from the request cookies and decrypts them.
func (o *Verifier) getCookies(r *http.Request) (sessionIDPtr *uuid.UUID, idToken string, refreshToken string, err error) {
	sessionIDCookie, err := r.Cookie(cookieNameSessionID)
	if err != nil && !errors.Is(err, http.ErrNoCookie) {
		return nil, "", "", fmt.Errorf("Failed to get session ID cookie from request: %w", err)
	}

	if sessionIDCookie == nil {
		return nil, "", "", nil
	}

	sessionID, err := uuid.Parse(sessionIDCookie.Value)
	if err != nil {
		return nil, "", "", fmt.Errorf("Invalid session ID cookie: %w", err)
	}

	secureCookie, err := o.secureCookieFromSession(sessionID)
	if err != nil {
		return nil, "", "", fmt.Errorf("Failed to decrypt cookies: %w", err)
	}

	idTokenCookie, err := r.Cookie(cookieNameIDToken)
	if err != nil && !errors.Is(err, http.ErrNoCookie) {
		return nil, "", "", fmt.Errorf("Failed to get ID token cookie from request: %w", err)
	}

	if idTokenCookie != nil {
		err = secureCookie.Decode(cookieNameIDToken, idTokenCookie.Value, &idToken)
		if err != nil {
			return nil, "", "", fmt.Errorf("Failed to decrypt ID token cookie: %w", err)
		}
	}

	refreshTokenCookie, err := r.Cookie(cookieNameRefreshToken)
	if err != nil && !errors.Is(err, http.ErrNoCookie) {
		return nil, "", "", fmt.Errorf("Failed to get refresh token cookie from request: %w", err)
	}

	if refreshTokenCookie != nil {
		err = secureCookie.Decode(cookieNameRefreshToken, refreshTokenCookie.Value, &refreshToken)
		if err != nil {
			return nil, "", "", fmt.Errorf("Failed to decrypt refresh token cookie: %w", err)
		}
	}

	return &sessionID, idToken, refreshToken, nil
}

// setCookies encrypts the session, ID, and refresh tokens and sets them in the HTTP response. Cookies are only set if they are
// non-empty. If isDelete is true, the values are set to empty strings and the cookie expiry is set to unix zero time.
func (*Verifier) setCookies(w http.ResponseWriter, secureCookie *securecookie.SecureCookie, sessionID uuid.UUID, idToken string, refreshToken string, isDelete bool) error {
	idTokenCookie := http.Cookie{
		Name:     cookieNameIDToken,
		Path:     "/",
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	}

	refreshTokenCookie := http.Cookie{
		Name:     cookieNameRefreshToken,
		Path:     "/",
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	}

	sessionIDCookie := http.Cookie{
		Name:     cookieNameSessionID,
		Path:     "/",
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	}

	if isDelete {
		idTokenCookie.Expires = time.Unix(0, 0)
		refreshTokenCookie.Expires = time.Unix(0, 0)
		sessionIDCookie.Expires = time.Unix(0, 0)

		http.SetCookie(w, &idTokenCookie)
		http.SetCookie(w, &refreshTokenCookie)
		http.SetCookie(w, &sessionIDCookie)
		return nil
	}

	encodedIDTokenCookie, err := secureCookie.Encode(cookieNameIDToken, idToken)
	if err != nil {
		return fmt.Errorf("Failed to encrypt ID token: %w", err)
	}

	encodedRefreshToken, err := secureCookie.Encode(cookieNameRefreshToken, refreshToken)
	if err != nil {
		return fmt.Errorf("Failed to encrypt refresh token: %w", err)
	}

	sessionIDCookie.Value = sessionID.String()
	idTokenCookie.Value = encodedIDTokenCookie
	refreshTokenCookie.Value = encodedRefreshToken

	http.SetCookie(w, &idTokenCookie)
	http.SetCookie(w, &refreshTokenCookie)
	http.SetCookie(w, &sessionIDCookie)
	return nil
}

// secureCookieFromSession returns a *securecookie.SecureCookie that is secure, unique to each client, and possible to
// decrypt on all cluster members.
//
// To do this we use the cluster private key as an input seed to HKDF (https://datatracker.ietf.org/doc/html/rfc5869) and
// use the given sessionID uuid.UUID as a salt. The session ID can then be stored as a plaintext cookie so that we can
// regenerate the keys upon the next request.
//
// Warning: Changes to this function might cause all existing OIDC users to be logged out of LXD (but not logged out of
// the IdP).
func (o *Verifier) secureCookieFromSession(sessionID uuid.UUID) (*securecookie.SecureCookie, error) {
	// Get the sessionID as a binary so that we can use it as a salt.
	salt, err := sessionID.MarshalBinary()
	if err != nil {
		return nil, fmt.Errorf("Failed to marshal session ID as binary: %w", err)
	}

	// Get the current cluster private key.
	clusterPrivateKey := o.clusterCert().PrivateKey()

	// Extract a pseudo-random key from the cluster private key.
	prk := hkdf.Extract(sha512.New, clusterPrivateKey, salt)

	// Get an io.Reader from which we can read a secure key. We will use this key as the hash key for the cookie.
	// The hash key is used to verify the integrity of decrypted values using HMAC. The HKDF "info" is set to "INTEGRITY"
	// to indicate the intended usage of the key and prevent decryption in other contexts
	// (see https://datatracker.ietf.org/doc/html/rfc5869#section-3.2).
	keyDerivationFunc := hkdf.Expand(sha512.New, prk, []byte("INTEGRITY"))

	// Read 64 bytes of the derived key. The securecookie library recommends 64 bytes for the hash key (https://github.com/gorilla/securecookie).
	cookieHashKey := make([]byte, 64)
	_, err = io.ReadFull(keyDerivationFunc, cookieHashKey)
	if err != nil {
		return nil, fmt.Errorf("Failed creating secure cookie hash key: %w", err)
	}

	// Get an io.Reader from which we can read a secure key. We will use this key as the block key for the cookie.
	// The block key is used by securecookie to perform AES encryption. The HKDF "info" is set to "ENCRYPTION"
	// to indicate the intended usage of the key and prevent decryption in other contexts
	// (see https://datatracker.ietf.org/doc/html/rfc5869#section-3.2).
	keyDerivationFunc = hkdf.Expand(sha512.New, prk, []byte("ENCRYPTION"))

	// Read 32 bytes of the derived key. Given 32 bytes for the block key the securecookie library will use AES-256 for encryption.
	cookieBlockKey := make([]byte, 32)
	_, err = io.ReadFull(keyDerivationFunc, cookieBlockKey)
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
func NewVerifier(issuer string, clientID string, clientSecret string, audience string, cert *shared.CertInfo) (*Verifier, error) {
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
	}

	return verifier, nil
}

func getCallbackURL(host string) string {
	return fmt.Sprintf("https://%s/oidc/callback", host)
}

func setUserInfoInRequest(authResult *AuthenticationResult, r *http.Request) {
	userInfo := &types.UserInfo{
		Email: authResult.Email,
		Name:  authResult.Name,
	}

	userInfoCtx := context.WithValue(r.Context(), types.UserInfoKey, userInfo)
	*r = *r.WithContext(userInfoCtx)
}

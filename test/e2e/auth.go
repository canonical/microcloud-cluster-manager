package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/canonical/lxd/shared/api"
	"github.com/canonical/microcloud-cluster-manager/test/helpers"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// testAuthAdminUserAllowsAccess verifies that a fully authenticated session with admin group
// can access a protected endpoint.
func testAuthAdminUserAllowsAccess(env *helpers.Environment) (testName string, testFunc func(t *testing.T)) {
	return "admin user is granted access to authenticated endpoint", func(t *testing.T) {
		const condition = "Should return 200 when admin session cookies are present"
		cookies, err := helpers.GetCookies(env, "cluster-manager-e2e-tests@example.org", "cluster-manager-e2e-password")
		if err != nil {
			helpers.LogTestOutcome(t, condition, err)
			return
		}

		path := api.NewURL().Scheme("https").Host(env.ManagementAPIHostPort()).Path("1.0", "remote-cluster")
		statusCode, err := helpers.QueryManagementAPI(env, http.MethodGet, path, nil, nil, helpers.AddCookiesToRequest(cookies))
		if err != nil {
			helpers.LogTestOutcome(t, condition, fmt.Errorf("request error: %w", err))
			return
		}

		if statusCode != http.StatusOK {
			helpers.LogTestOutcome(t, condition, fmt.Errorf("expected 200, got %d", statusCode))
			return
		}

		helpers.LogTestOutcome(t, condition, nil)
	}
}

// testAuthNonAdminDenyAccess verifies that a non-admin user is denied access to a protected endpoint.
func testAuthNonAdminDenyAccess(env *helpers.Environment) (testName string, testFunc func(t *testing.T)) {
	return "non-admin user is denied access to authenticated endpoint", func(t *testing.T) {
		const condition = "Should return 403 when non-admin session cookies are present"
		cookies, err := helpers.GetCookies(env, "cluster-manager-e2e-lower-permission@example.org", "cluster-manager-e2e-password")
		if err != nil {
			helpers.LogTestOutcome(t, condition, err)
			return
		}

		path := api.NewURL().Scheme("https").Host(env.ManagementAPIHostPort()).Path("1.0", "remote-cluster")
		statusCode, err := helpers.QueryManagementAPI(env, http.MethodGet, path, nil, nil, helpers.AddCookiesToRequest(cookies))

		if err != nil && api.StatusErrorCheck(err, http.StatusForbidden) {
			helpers.LogTestOutcome(t, condition, nil)
			return
		}

		helpers.LogTestOutcome(t, condition, fmt.Errorf("expected 403 got %d: %w", statusCode, err))
	}
}

// testAuthNonAdminUnprotectedEndpointAllowAccess verifies that a non-admin user is granted access to an unprotected endpoint.
func testAuthNonAdminUnprotectedEndpointAllowAccess(env *helpers.Environment) (testName string, testFunc func(t *testing.T)) {
	return "non-admin user is granted access to unprotected endpoint", func(t *testing.T) {
		const condition = "Should return 200 when non-admin session cookies are present on unprotected endpoint"
		cookies, err := helpers.GetCookies(env, "cluster-manager-e2e-lower-permission@example.org", "cluster-manager-e2e-password")
		if err != nil {
			helpers.LogTestOutcome(t, condition, err)
			return
		}

		path := api.NewURL().Scheme("https").Host(env.ManagementAPIHostPort()).Path("1.0", "status")
		statusCode, err := helpers.QueryManagementAPI(env, http.MethodGet, path, nil, nil, helpers.AddCookiesToRequest(cookies))

		if err != nil {
			helpers.LogTestOutcome(t, condition, fmt.Errorf("request error: %w", err))
			return
		}

		if statusCode != http.StatusOK {
			helpers.LogTestOutcome(t, condition, fmt.Errorf("expected 200, got %d", statusCode))
			return
		}

		helpers.LogTestOutcome(t, condition, nil)
	}
}

// testAuthTamperedSessionTokenDeniesAccess verifies that a request with a tampered
// session token cookie (invalid signature) is rejected.
func testAuthTamperedSessionTokenDeniesAccess(env *helpers.Environment) (testName string, testFunc func(t *testing.T)) {
	return "tampered session token cookie is rejected", func(t *testing.T) {
		const condition = "Should return 401 when the session token cookie has been tampered with"
		cookies, err := helpers.GetCookies(env, "cluster-manager-e2e-tests@example.org", "cluster-manager-e2e-password")
		if err != nil {
			helpers.LogTestOutcome(t, condition, err)
			return
		}

		// Replace the session token with a different value while keeping session ID intact.
		tampered := make([]*http.Cookie, len(cookies))
		for i, c := range cookies {
			cloned := *c
			if cloned.Name == "oidc_session" {
				cloned.Value = "tampered.invalid.cookie.value"
			}

			tampered[i] = &cloned
		}

		path := api.NewURL().Scheme("https").Host(env.ManagementAPIHostPort()).Path("1.0", "remote-cluster")
		statusCode, err := helpers.QueryManagementAPI(env, http.MethodGet, path, nil, nil, helpers.AddCookiesToRequest(tampered))

		if err != nil && api.StatusErrorCheck(err, http.StatusUnauthorized) {
			helpers.LogTestOutcome(t, condition, nil)
			return
		}
		helpers.LogTestOutcome(t, condition, fmt.Errorf("expected 401, got %d", statusCode))
	}
}

// testAuthMissingUserSecretCookieDeniesAccess verifies that a request with a valid oidc_session
// cookie but no user_secret cookie is rejected. The server needs the user_secret to decrypt
// the stored OIDC tokens for the session.
func testAuthMissingUserSecretCookieDeniesAccess(env *helpers.Environment) (testName string, testFunc func(t *testing.T)) {
	return "request with valid oidc_session but missing user_secret cookie is rejected", func(t *testing.T) {
		const condition = "Should return 401 when user_secret cookie is absent"
		cookies, err := helpers.GetCookies(env, "cluster-manager-e2e-tests@example.org", "cluster-manager-e2e-password")
		if err != nil {
			helpers.LogTestOutcome(t, condition, err)
			return
		}

		// Keep every cookie except user_secret.
		stripped := make([]*http.Cookie, 0, len(cookies))
		for _, c := range cookies {
			if c.Name != "user_secret" {
				stripped = append(stripped, c)
			}
		}

		path := api.NewURL().Scheme("https").Host(env.ManagementAPIHostPort()).Path("1.0", "remote-cluster")
		statusCode, err := helpers.QueryManagementAPI(env, http.MethodGet, path, nil, nil, helpers.AddCookiesToRequest(stripped))

		if err != nil && api.StatusErrorCheck(err, http.StatusUnauthorized) {
			helpers.LogTestOutcome(t, condition, nil)
			return
		}

		helpers.LogTestOutcome(t, condition, fmt.Errorf("expected 401, got %d", statusCode))
	}
}

// testAuthMissingSessionCookieDeniesAccess verifies that a request carrying the user_secret
// cookie but no oidc_session cookie is rejected.
func testAuthMissingSessionCookieDeniesAccess(env *helpers.Environment) (testName string, testFunc func(t *testing.T)) {
	return "request with user_secret but missing oidc_session cookie is rejected", func(t *testing.T) {
		const condition = "Should return 401 when oidc_session cookie is absent"
		cookies, err := helpers.GetCookies(env, "cluster-manager-e2e-tests@example.org", "cluster-manager-e2e-password")
		if err != nil {
			helpers.LogTestOutcome(t, condition, err)
			return
		}

		// Keep every cookie except oidc_session.
		stripped := make([]*http.Cookie, 0, len(cookies))
		for _, c := range cookies {
			if c.Name != "oidc_session" {
				stripped = append(stripped, c)
			}
		}

		path := api.NewURL().Scheme("https").Host(env.ManagementAPIHostPort()).Path("1.0", "remote-cluster")
		statusCode, err := helpers.QueryManagementAPI(env, http.MethodGet, path, nil, nil, helpers.AddCookiesToRequest(stripped))

		if err != nil && api.StatusErrorCheck(err, http.StatusUnauthorized) {
			helpers.LogTestOutcome(t, condition, nil)
			return
		}

		helpers.LogTestOutcome(t, condition, fmt.Errorf("expected 401, got %d", statusCode))
	}
}

// testAuthExpiredSessionMissingUserSecretDeniesAccess verifies that an expired session token
// with a correct signature but no user_secret cookie is rejected. The session renewal path
// also requires the user_secret to decrypt the stored refresh token.
func testAuthExpiredSessionMissingUserSecretDeniesAccess(env *helpers.Environment) (testName string, testFunc func(t *testing.T)) {
	return "expired session token with correct signature but missing user_secret cookie is rejected", func(t *testing.T) {
		const condition = "Should return 401 when session token is expired and user_secret cookie is absent"
		cookies, err := helpers.GetCookies(env, "cluster-manager-e2e-tests@example.org", "cluster-manager-e2e-password")
		if err != nil {
			helpers.LogTestOutcome(t, condition, err)
			return
		}

		var sessionToken string
		for _, c := range cookies {
			if c.Name == "oidc_session" {
				sessionToken = c.Value
				break
			}
		}

		if sessionToken == "" {
			helpers.LogTestOutcome(t, condition, fmt.Errorf("oidc_session cookie not found after login"))
			return
		}

		parsedClaims := &jwt.RegisteredClaims{}
		_, _, err = jwt.NewParser().ParseUnverified(sessionToken, parsedClaims)
		if err != nil {
			helpers.LogTestOutcome(t, condition, fmt.Errorf("failed to parse session JWT: %w", err))
			return
		}

		sessionID, err := uuid.Parse(parsedClaims.Subject)
		if err != nil {
			helpers.LogTestOutcome(t, condition, fmt.Errorf("failed to parse session ID from JWT subject: %w", err))
			return
		}

		signingKey, err := helpers.DeriveSessionSigningKey(sessionID, env.ManagementAPICert().PrivateKey())
		if err != nil {
			helpers.LogTestOutcome(t, condition, err)
			return
		}

		expiredSigned, err := helpers.CreateExpiredSessionToken(sessionID, signingKey)
		if err != nil {
			helpers.LogTestOutcome(t, condition, err)
			return
		}

		// Replace oidc_session with the expired token and drop user_secret entirely.
		modified := make([]*http.Cookie, 0, len(cookies))
		for _, c := range cookies {
			if c.Name == "user_secret" {
				continue
			}
			cloned := *c
			if cloned.Name == "oidc_session" {
				cloned.Value = expiredSigned
			}
			modified = append(modified, &cloned)
		}

		path := api.NewURL().Scheme("https").Host(env.ManagementAPIHostPort()).Path("1.0", "remote-cluster")
		statusCode, err := helpers.QueryManagementAPI(env, http.MethodGet, path, nil, nil, helpers.AddCookiesToRequest(modified))

		if err != nil && api.StatusErrorCheck(err, http.StatusUnauthorized) {
			helpers.LogTestOutcome(t, condition, nil)
			return
		}

		helpers.LogTestOutcome(t, condition, fmt.Errorf("expected 401, got %d", statusCode))
	}
}

// testAuthOIDCLoginRedirects verifies that the /oidc/login endpoint responds with
// a redirect to the identity provider.
func testAuthOIDCLoginRedirects(env *helpers.Environment) (testName string, testFunc func(t *testing.T)) {
	return "/oidc/login redirects to the identity provider", func(t *testing.T) {
		condition := "Should redirect to the OIDC identity provider"

		certPublicKey, err := env.ManagementAPICert().PublicKeyX509()
		if err != nil {
			helpers.LogTestOutcome(t, condition, fmt.Errorf("failed to get cert: %w", err))
			return
		}

		tlsClient, err := helpers.NewTLSHTTPClient(api.URL{}, nil, certPublicKey, env.ManagementAPIHost())
		if err != nil {
			helpers.LogTestOutcome(t, condition, fmt.Errorf("failed to create TLS client: %w", err))
			return
		}

		// Do not follow redirects so we can inspect the Location header.
		tlsClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}

		loginURL := fmt.Sprintf("https://%s/oidc/login", env.ManagementAPIHostPort())

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, loginURL, nil)
		if err != nil {
			helpers.LogTestOutcome(t, condition, fmt.Errorf("failed to build request: %w", err))
			return
		}

		res, err := tlsClient.Do(req)
		if err != nil && !errors.Is(err, http.ErrUseLastResponse) {
			helpers.LogTestOutcome(t, condition, fmt.Errorf("request failed: %w", err))
			return
		}

		defer res.Body.Close()

		if res.StatusCode != http.StatusFound && res.StatusCode != http.StatusMovedPermanently && res.StatusCode != http.StatusSeeOther {
			helpers.LogTestOutcome(t, condition, fmt.Errorf("expected a redirect (3xx), got %d", res.StatusCode))
			return
		}

		location := res.Header.Get("Location")
		if location == "" {
			helpers.LogTestOutcome(t, condition, fmt.Errorf("expected a Location header in redirect response, got none"))
			return
		}

		helpers.LogTestOutcome(t, condition, nil)
	}
}

// testAuthExpiredSessionCorrectSignatureAllowsAccess verifies that a session token that is expired
// but bears a correct signature triggers transparent session renewal and still grants access,
// provided the refresh token stored in the session is still valid.
func testAuthExpiredSessionCorrectSignatureAllowsAccess(env *helpers.Environment) (testName string, testFunc func(t *testing.T)) {
	return "expired session token with correct signature and valid refresh token triggers transparent renewal", func(t *testing.T) {
		// NOTE: The server only permits access when the session token is expired with a correct signature AND
		// the underlying refresh token is still valid. A correctly signed but expired token alone is not
		// sufficient — if the refresh token were also expired, the server would return 401.
		// This test relies on a freshly issued session so that the refresh token is guaranteed to be valid.
		const condition = "Should return 200 when session token is expired but correctly signed and the refresh token is still valid"
		cookies, err := helpers.GetCookies(env, "cluster-manager-e2e-tests@example.org", "cluster-manager-e2e-password")
		if err != nil {
			helpers.LogTestOutcome(t, condition, err)
			return
		}

		// Find the oidc_session cookie that contains the signed session JWT.
		var sessionToken string
		for _, c := range cookies {
			if c.Name == "oidc_session" {
				sessionToken = c.Value
				break
			}
		}

		if sessionToken == "" {
			helpers.LogTestOutcome(t, condition, fmt.Errorf("oidc_session cookie not found after login"))
			return
		}

		// Parse the JWT without verification to extract the session ID from the sub claim.
		parsedClaims := &jwt.RegisteredClaims{}
		_, _, err = jwt.NewParser().ParseUnverified(sessionToken, parsedClaims)
		if err != nil {
			helpers.LogTestOutcome(t, condition, fmt.Errorf("failed to parse session JWT: %w", err))
			return
		}

		sessionID, err := uuid.Parse(parsedClaims.Subject)
		if err != nil {
			helpers.LogTestOutcome(t, condition, fmt.Errorf("failed to parse session ID from JWT subject: %w", err))
			return
		}

		signingKey, err := helpers.DeriveSessionSigningKey(sessionID, env.ManagementAPICert().PrivateKey())
		if err != nil {
			helpers.LogTestOutcome(t, condition, err)
			return
		}

		expiredTokenSigned, err := helpers.CreateExpiredSessionToken(sessionID, signingKey)
		if err != nil {
			helpers.LogTestOutcome(t, condition, err)
			return
		}

		// Swap the oidc_session cookie for the expired-but-correctly-signed token.
		modifiedCookies := make([]*http.Cookie, len(cookies))
		for i, c := range cookies {
			cloned := *c
			if cloned.Name == "oidc_session" {
				cloned.Value = expiredTokenSigned
			}
			modifiedCookies[i] = &cloned
		}

		path := api.NewURL().Scheme("https").Host(env.ManagementAPIHostPort()).Path("1.0", "remote-cluster")
		statusCode, err := helpers.QueryManagementAPI(env, http.MethodGet, path, nil, nil, helpers.AddCookiesToRequest(modifiedCookies))
		if err != nil {
			helpers.LogTestOutcome(t, condition, fmt.Errorf("request error: %w", err))
			return
		}

		if statusCode != http.StatusOK {
			helpers.LogTestOutcome(t, condition, fmt.Errorf("expected 200, got %d", statusCode))
			return
		}

		helpers.LogTestOutcome(t, condition, nil)
	}
}

// testAuthTamperedUserSecretCookieDeniesAccess verifies that a request with a valid oidc_session
// but a tampered user_secret cookie is rejected. The server decodes user_secret using a per-session
// securecookie key, so any modification invalidates the MAC.
func testAuthTamperedUserSecretCookieDeniesAccess(env *helpers.Environment) (testName string, testFunc func(t *testing.T)) {
	return "tampered user_secret cookie is rejected", func(t *testing.T) {
		const condition = "Should return 401 when user_secret cookie has been tampered with"
		cookies, err := helpers.GetCookies(env, "cluster-manager-e2e-tests@example.org", "cluster-manager-e2e-password")
		if err != nil {
			helpers.LogTestOutcome(t, condition, err)
			return
		}

		tampered := make([]*http.Cookie, len(cookies))
		for i, c := range cookies {
			cloned := *c
			if cloned.Name == "user_secret" {
				cloned.Value = "tampered.invalid.user.secret.value"
			}
			tampered[i] = &cloned
		}

		path := api.NewURL().Scheme("https").Host(env.ManagementAPIHostPort()).Path("1.0", "remote-cluster")
		statusCode, err := helpers.QueryManagementAPI(env, http.MethodGet, path, nil, nil, helpers.AddCookiesToRequest(tampered))

		if err != nil && api.StatusErrorCheck(err, http.StatusUnauthorized) {
			helpers.LogTestOutcome(t, condition, nil)
			return
		}

		helpers.LogTestOutcome(t, condition, fmt.Errorf("expected 401, got %d", statusCode))
	}
}

// testAuthValidSignatureUnknownSessionIDDeniesAccess verifies that a session token signed
// with the correct HKDF key for a freshly generated session ID (one that does not exist in the
// database) is rejected. The server should fail the database lookup and return 401.
func testAuthValidSignatureUnknownSessionIDDeniesAccess(env *helpers.Environment) (testName string, testFunc func(t *testing.T)) {
	return "correctly signed session token for unknown session ID is rejected", func(t *testing.T) {
		const condition = "Should return 401 when session ID is not found in the database"
		cookies, err := helpers.GetCookies(env, "cluster-manager-e2e-tests@example.org", "cluster-manager-e2e-password")
		if err != nil {
			helpers.LogTestOutcome(t, condition, err)
			return
		}

		// Generate a fresh session ID that has never been stored in the database.
		unknownSessionID, err := uuid.NewV7()
		if err != nil {
			helpers.LogTestOutcome(t, condition, fmt.Errorf("failed to generate session ID: %w", err))
			return
		}

		signingKey, err := helpers.DeriveSessionSigningKey(unknownSessionID, env.ManagementAPICert().PrivateKey())
		if err != nil {
			helpers.LogTestOutcome(t, condition, err)
			return
		}

		claims := &jwt.RegisteredClaims{
			Subject:   unknownSessionID.String(),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24 * 7)),
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
		signedToken, err := token.SignedString(signingKey)
		if err != nil {
			helpers.LogTestOutcome(t, condition, fmt.Errorf("failed to sign session JWT: %w", err))
			return
		}

		// Replace only the oidc_session cookie; keep user_secret so cookie parsing proceeds.
		modified := make([]*http.Cookie, len(cookies))
		for i, c := range cookies {
			cloned := *c
			if cloned.Name == "oidc_session" {
				cloned.Value = signedToken
			}
			modified[i] = &cloned
		}

		path := api.NewURL().Scheme("https").Host(env.ManagementAPIHostPort()).Path("1.0", "remote-cluster")
		statusCode, err := helpers.QueryManagementAPI(env, http.MethodGet, path, nil, nil, helpers.AddCookiesToRequest(modified))

		if err != nil && api.StatusErrorCheck(err, http.StatusUnauthorized) {
			helpers.LogTestOutcome(t, condition, nil)
			return
		}

		helpers.LogTestOutcome(t, condition, fmt.Errorf("expected 401, got %d", statusCode))
	}
}

// testAuthLoggedOutSessionDeniesAccess verifies that cookies from a session that has been
// logged out are rejected on subsequent requests. After logout the session is deleted from
// the database, so replaying the original cookies must return 401.
func testAuthLoggedOutSessionDeniesAccess(env *helpers.Environment) (testName string, testFunc func(t *testing.T)) {
	return "cookies from a logged-out session are rejected", func(t *testing.T) {
		const condition = "Should return 401 when cookies belong to an already logged-out session"
		cookies, err := helpers.GetCookies(env, "cluster-manager-e2e-tests@example.org", "cluster-manager-e2e-password")
		if err != nil {
			helpers.LogTestOutcome(t, condition, err)
			return
		}

		// Confirm the session is valid before logging out.
		path := api.NewURL().Scheme("https").Host(env.ManagementAPIHostPort()).Path("1.0", "remote-cluster")
		statusCode, err := helpers.QueryManagementAPI(env, http.MethodGet, path, nil, nil, helpers.AddCookiesToRequest(cookies))
		if err != nil {
			helpers.LogTestOutcome(t, condition, fmt.Errorf("pre-logout request error: %w", err))
			return
		}

		if statusCode != http.StatusOK {
			helpers.LogTestOutcome(t, condition, fmt.Errorf("expected 200 before logout, got %d", statusCode))
			return
		}

		// Log out, which deletes the session from the database.
		err = helpers.LogoutFromManagementAPI(env, cookies)
		if err != nil {
			helpers.LogTestOutcome(t, condition, fmt.Errorf("logout failed: %w", err))
			return
		}

		// Replay the original cookies — the session no longer exists server-side.
		statusCode, err = helpers.QueryManagementAPI(env, http.MethodGet, path, nil, nil, helpers.AddCookiesToRequest(cookies))

		if err != nil && api.StatusErrorCheck(err, http.StatusUnauthorized) {
			helpers.LogTestOutcome(t, condition, nil)
			return
		}

		helpers.LogTestOutcome(t, condition, fmt.Errorf("expected 401 after logout, got %d: %w", statusCode, err))
	}
}

// testAuthOIDCCallbackWithInvalidStateReturns403 verifies that hitting the callback
// endpoint with a malformed state parameter returns an error response.
func testAuthOIDCCallbackWithInvalidStateReturns403(env *helpers.Environment) (testName string, testFunc func(t *testing.T)) {
	return "/oidc/callback with malformed state returns an error", func(t *testing.T) {
		condition := "Should return an error response for a malformed state parameter"

		oidcCallbackURL := api.NewURL().Scheme("https").Host(env.ManagementAPIHostPort()).Path("oidc", "callback")
		oidcCallbackURL.RawQuery = "state=not-valid-base64%%21%%21&code=fakecode"
		statusCode, err := helpers.QueryManagementAPI(env, http.MethodGet, oidcCallbackURL, nil, nil, nil)
		if err != nil && statusCode == http.StatusForbidden {
			helpers.LogTestOutcome(t, condition, nil)
			return
		}
		helpers.LogTestOutcome(t, condition, fmt.Errorf("expected a 403 error response for invalid state, got %d", statusCode))
	}
}

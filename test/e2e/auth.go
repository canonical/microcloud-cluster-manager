package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/canonical/lxd/shared/api"
	"github.com/canonical/microcloud-cluster-manager/test/helpers"
	"github.com/google/uuid"
)

// testAuthAdminUserAllowsAccess verifies that a fully authenticated session with admin group
// can access a protected endpoint.
func testAuthAdminUserAllowsAccess(env *helpers.Environment) (testName string, testFunc func(t *testing.T)) {
	return "admin user is granted access to authenticated endpoint", func(t *testing.T) {
		const condition = "Should return 200 when admin session cookies are present"
		cookies, err := helpers.GetValidCookies(env, "cluster-manager-e2e-tests@example.org", "cluster-manager-e2e-password")
		if err != nil {
			helpers.LogTestOutcome(t, condition, err)
			return
		}

		statusCode, err := helpers.DoRequestWithCookies(env, "/1.0/remote-cluster", http.MethodGet, cookies)
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
		cookies, err := helpers.GetValidCookies(env, "cluster-manager-e2e-lower-permission@example.org", "cluster-manager-e2e-password")
		if err != nil {
			helpers.LogTestOutcome(t, condition, err)
			return
		}

		statusCode, err := helpers.DoRequestWithCookies(env, "/1.0/remote-cluster", http.MethodGet, cookies)
		if err != nil {
			helpers.LogTestOutcome(t, condition, fmt.Errorf("request error: %w", err))
			return
		}

		if statusCode != http.StatusForbidden {
			helpers.LogTestOutcome(t, condition, fmt.Errorf("expected 403, got %d", statusCode))
			return
		}

		helpers.LogTestOutcome(t, condition, nil)
	}
}

// testAuthNonAdminUnprotectedEndpointAllowAccess verifies that a non-admin user is granted access to an unprotected endpoint.
func testAuthNonAdminUnprotectedEndpointAllowAccess(env *helpers.Environment) (testName string, testFunc func(t *testing.T)) {
	return "non-admin user is granted access to unprotected endpoint", func(t *testing.T) {
		const condition = "Should return 200 when non-admin session cookies are present"
		cookies, err := helpers.GetValidCookies(env, "cluster-manager-e2e-lower-permission@example.org", "cluster-manager-e2e-password")
		if err != nil {
			helpers.LogTestOutcome(t, condition, err)
			return
		}

		statusCode, err := helpers.DoRequestWithCookies(env, "/1.0/status", http.MethodGet, cookies)
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

// testAuthNoCookiesDeniesAccess verifies that requests with no session cookies are rejected.
func testAuthNoCookiesDeniesAccess(env *helpers.Environment) (testName string, testFunc func(t *testing.T)) {
	return "requests with no cookies are rejected", func(t *testing.T) {
		const condition = "Should return 403 when no cookies are present"

		statusCode, err := helpers.DoRequestWithCookies(env, "/1.0/remote-cluster", http.MethodGet, nil)
		if err != nil {
			helpers.LogTestOutcome(t, condition, fmt.Errorf("request error: %w", err))
			return
		}

		if statusCode != http.StatusForbidden {
			helpers.LogTestOutcome(t, condition, fmt.Errorf("expected 403, got %d", statusCode))
			return
		}

		helpers.LogTestOutcome(t, condition, nil)
	}
}

// testAuthTamperedIDTokenDeniesAccess verifies that a request with a tampered
// ID token cookie (invalid signature) is rejected even when the session ID and
// refresh token cookies are present.
func testAuthTamperedIDTokenDeniesAccess(env *helpers.Environment) (testName string, testFunc func(t *testing.T)) {
	return "tampered ID token cookie is rejected", func(t *testing.T) {
		const condition = "Should return 403 when the ID token cookie has been tampered with"
		cookies, err := helpers.GetValidCookies(env, "cluster-manager-e2e-tests@example.org", "cluster-manager-e2e-password")
		if err != nil {
			helpers.LogTestOutcome(t, condition, err)
			return
		}

		// Replace the encrypted ID token with a different value while keeping session ID intact.
		tampered := make([]*http.Cookie, len(cookies))
		for i, c := range cookies {
			cloned := *c
			if cloned.Name == "oidc_identity" {
				cloned.Value = "tampered.invalid.cookie.value"
			}

			tampered[i] = &cloned
		}

		statusCode, err := helpers.DoRequestWithCookies(env, "/1.0/remote-cluster", http.MethodGet, tampered)
		if err != nil {
			helpers.LogTestOutcome(t, condition, fmt.Errorf("request error: %w", err))
			return
		}

		if statusCode != http.StatusForbidden {
			helpers.LogTestOutcome(t, condition, fmt.Errorf("expected 403, got %d", statusCode))
			return
		}

		helpers.LogTestOutcome(t, condition, nil)
	}
}

// testAuthTamperedRefreshTokenDeniesAccess verifies that a request where both the
// ID token and the refresh token have been tampered with is rejected.
// This forces the refresh path in authenticateIDToken by providing a tampered ID token.
func testAuthTamperedRefreshTokenDeniesAccess(env *helpers.Environment) (testName string, testFunc func(t *testing.T)) {
	return "tampered refresh token cookie is rejected", func(t *testing.T) {
		const condition = "Should return 403 when both the ID token and refresh token cookies are tampered"

		cookies, err := helpers.GetValidCookies(env, "cluster-manager-e2e-tests@example.org", "cluster-manager-e2e-password")
		if err != nil {
			helpers.LogTestOutcome(t, condition, err)
			return
		}

		tampered := make([]*http.Cookie, len(cookies))
		for i, c := range cookies {
			cloned := *c
			if cloned.Name == "oidc_identity" || cloned.Name == "oidc_refresh" {
				cloned.Value = "tampered.invalid.cookie.value"
			}

			tampered[i] = &cloned
		}

		statusCode, err := helpers.DoRequestWithCookies(env, "/1.0/remote-cluster", http.MethodGet, tampered)
		if err != nil {
			helpers.LogTestOutcome(t, condition, fmt.Errorf("request error: %w", err))
			return
		}

		if statusCode != http.StatusForbidden {
			helpers.LogTestOutcome(t, condition, fmt.Errorf("expected 403, got %d", statusCode))
			return
		}

		helpers.LogTestOutcome(t, condition, nil)
	}
}

// testAuthUnknownSessionIDDeniesAccess verifies that a session ID that does not
// match any known session (i.e. generates a different HKDF-derived cookie key) causes
// cookie decryption to fail and the request to be rejected.
func testAuthUnknownSessionIDDeniesAccess(env *helpers.Environment) (testName string, testFunc func(t *testing.T)) {
	return "unknown session ID causes decryption failure and is rejected", func(t *testing.T) {
		const condition = "Should return 403 when the session ID is a valid UUID but unknown to the server"
		cookies, err := helpers.GetValidCookies(env, "cluster-manager-e2e-tests@example.org", "cluster-manager-e2e-password")
		if err != nil {
			helpers.LogTestOutcome(t, condition, err)
			return
		}

		// Replace the session ID with a randomly generated UUID that was never issued.
		unknownSessionID := uuid.New().String()
		tampered := make([]*http.Cookie, len(cookies))
		for i, c := range cookies {
			cloned := *c
			if cloned.Name == "session_id" {
				cloned.Value = unknownSessionID
			}

			tampered[i] = &cloned
		}

		statusCode, err := helpers.DoRequestWithCookies(env, "/1.0/remote-cluster", http.MethodGet, tampered)
		if err != nil {
			helpers.LogTestOutcome(t, condition, fmt.Errorf("request error: %w", err))
			return
		}

		if statusCode != http.StatusForbidden {
			helpers.LogTestOutcome(t, condition, fmt.Errorf("expected 403, got %d", statusCode))
			return
		}

		helpers.LogTestOutcome(t, condition, nil)
	}
}

// testAuthOnlySessionIDCookieDeniesAccess verifies that presenting only a session ID
// cookie with no ID token or refresh token results in a rejection.
func testAuthOnlySessionIDCookieDeniesAccess(env *helpers.Environment) (testName string, testFunc func(t *testing.T)) {
	return "session ID cookie only (no tokens) is rejected", func(t *testing.T) {
		const condition = "Should return 403 when only the session ID cookie is present"

		// Use a freshly generated session ID so the server can parse it.
		cookies := []*http.Cookie{
			{Name: "session_id", Value: uuid.New().String()},
		}

		statusCode, err := helpers.DoRequestWithCookies(env, "/1.0/remote-cluster", http.MethodGet, cookies)
		if err != nil {
			helpers.LogTestOutcome(t, condition, fmt.Errorf("request error: %w", err))
			return
		}

		if statusCode != http.StatusForbidden {
			helpers.LogTestOutcome(t, condition, fmt.Errorf("expected 403, got %d", statusCode))
			return
		}

		helpers.LogTestOutcome(t, condition, nil)
	}
}

// testAuthLogoutInvalidatesCookies verifies that after a successful logout the
// cookies no longer grant access to a protected endpoint.
// func testAuthLogoutInvalidatesCookies(env *helpers.Environment) (testName string, testFunc func(t *testing.T)) {
// 	return "cookies are invalidated after logout", func(t *testing.T) {
// 		var condition string

// 		condition = "Should be able to log in and access authenticated endpoint before logout"
// 		cookies, err := helpers.GetValidCookies(env, "cluster-manager-e2e-tests@example.org", "cluster-manager-e2e-password")
// 		if err != nil {
// 			helpers.LogTestOutcome(t, condition, err)
// 			return
// 		}

// 		statusCode, err := helpers.DoRequestWithCookies(env, "/1.0/remote-cluster", http.MethodGet, cookies)
// 		if err != nil {
// 			helpers.LogTestOutcome(t, condition, fmt.Errorf("request error: %w", err))
// 			return
// 		}

// 		if statusCode != http.StatusOK {
// 			helpers.LogTestOutcome(t, condition, fmt.Errorf("expected 200 before logout, got %d", statusCode))
// 			return
// 		}

// 		helpers.LogTestOutcome(t, condition, nil)

// 		condition = "Should invalidate session cookies after calling /oidc/logout"
// 		certPublicKey, err := env.ManagementAPICert().PublicKeyX509()
// 		if err != nil {
// 			helpers.LogTestOutcome(t, condition, fmt.Errorf("failed to get cert: %w", err))
// 			return
// 		}

// 		tlsClient, err := helpers.NewTLSHTTPClient(api.URL{}, nil, certPublicKey, env.ManagementAPIHost())
// 		if err != nil {
// 			helpers.LogTestOutcome(t, condition, fmt.Errorf("failed to create TLS client: %w", err))
// 			return
// 		}

// 		logoutURL := fmt.Sprintf("https://%s/oidc/logout", env.ManagementAPIHostPort())
// 		logoutReq, err := http.NewRequestWithContext(context.Background(), http.MethodGet, logoutURL, nil)
// 		if err != nil {
// 			helpers.LogTestOutcome(t, condition, fmt.Errorf("failed to build logout request: %w", err))
// 			return
// 		}

// 		for _, c := range cookies {
// 			logoutReq.AddCookie(c)
// 		}

// 		// Do not follow redirects so we can inspect the Set-Cookie headers directly.
// 		tlsClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
// 			return http.ErrUseLastResponse
// 		}

// 		logoutResp, err := tlsClient.Do(logoutReq)
// 		if err != nil && !errors.Is(err, http.ErrUseLastResponse) {
// 			helpers.LogTestOutcome(t, condition, fmt.Errorf("logout request failed: %w", err))
// 			return
// 		}

// 		defer logoutResp.Body.Close()

// 		// Verify that the response expires the relevant cookies.
// 		expiredCookies := map[string]bool{}
// 		for _, sc := range logoutResp.Cookies() {
// 			if sc.Expires.Equal(time.Unix(0, 0)) || sc.MaxAge < 0 {
// 				expiredCookies[sc.Name] = true
// 			}
// 		}

// 		for _, name := range []string{"oidc_identity", "oidc_refresh", "session_id"} {
// 			if !expiredCookies[name] {
// 				helpers.LogTestOutcome(t, condition, fmt.Errorf("cookie %q was not expired by logout response", name))
// 				return
// 			}
// 		}

// 		helpers.LogTestOutcome(t, condition, nil)

// 		condition = "Should return 403 when using original cookies after logout"
// 		statusCode, err = helpers.DoRequestWithCookies(env, "/1.0/remote-cluster", http.MethodGet, cookies)
// 		if err != nil {
// 			helpers.LogTestOutcome(t, condition, fmt.Errorf("request error: %w", err))
// 			return
// 		}

// 		if statusCode != http.StatusForbidden {
// 			helpers.LogTestOutcome(t, condition, fmt.Errorf("expected 403 after logout, got %d", statusCode))
// 			return
// 		}

// 		helpers.LogTestOutcome(t, condition, nil)
// 	}
// }

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
		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, loginURL, nil)
		if err != nil {
			helpers.LogTestOutcome(t, condition, fmt.Errorf("failed to build request: %w", err))
			return
		}

		resp, err := tlsClient.Do(req)
		if err != nil && !errors.Is(err, http.ErrUseLastResponse) {
			helpers.LogTestOutcome(t, condition, fmt.Errorf("request failed: %w", err))
			return
		}

		defer resp.Body.Close()

		if resp.StatusCode != http.StatusFound && resp.StatusCode != http.StatusMovedPermanently && resp.StatusCode != http.StatusSeeOther {
			helpers.LogTestOutcome(t, condition, fmt.Errorf("expected a redirect (3xx), got %d", resp.StatusCode))
			return
		}

		location := resp.Header.Get("Location")
		if location == "" {
			helpers.LogTestOutcome(t, condition, fmt.Errorf("expected a Location header in redirect response, got none"))
			return
		}

		helpers.LogTestOutcome(t, condition, nil)
	}
}

// testAuthOIDCCallbackWithInvalidStateReturns500 verifies that hitting the callback
// endpoint with a malformed state parameter returns an error response.
func testAuthOIDCCallbackWithInvalidStateReturns500(env *helpers.Environment) (testName string, testFunc func(t *testing.T)) {
	return "/oidc/callback with malformed state returns an error", func(t *testing.T) {
		condition := "Should return an error response for a malformed state parameter"

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

		callbackURL := fmt.Sprintf("https://%s/oidc/callback?state=not-valid-base64%%21%%21&code=fakecode", env.ManagementAPIHostPort())
		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, callbackURL, nil)
		if err != nil {
			helpers.LogTestOutcome(t, condition, fmt.Errorf("failed to build request: %w", err))
			return
		}

		resp, err := tlsClient.Do(req)
		if err != nil {
			helpers.LogTestOutcome(t, condition, fmt.Errorf("request failed: %w", err))
			return
		}

		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			helpers.LogTestOutcome(t, condition, fmt.Errorf("expected a non-200 error response for invalid state, got 200"))
			return
		}

		helpers.LogTestOutcome(t, condition, nil)
	}
}

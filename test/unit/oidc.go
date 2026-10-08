package main

import (
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/canonical/microcloud-cluster-manager/test/helpers"
)

const loginIDCookieName = "login_id"

// findCookie returns the cookie with the given name, or nil if there is none.
func findCookie(cookies []*http.Cookie, name string) *http.Cookie {
	for _, cookie := range cookies {
		if cookie.Name == name {
			return cookie
		}
	}

	return nil
}

// checkCodeExchange checks that the IdP received a token exchange request whose PKCE code verifier matches the code
// challenge sent in the login redirect. This only happens if the callback could decrypt the state and PKCE cookies.
func checkCodeExchange(idp *helpers.MockIdentityProvider, loginRes *http.Response) error {
	location, err := loginRes.Location()
	if err != nil {
		return fmt.Errorf("failed to get login redirect location: %w", err)
	}

	codeVerifier := idp.CodeVerifier()
	if codeVerifier == "" {
		return errors.New("callback did not exchange the authorization code with the IdP")
	}

	hash := sha256.Sum256([]byte(codeVerifier))
	if base64.RawURLEncoding.EncodeToString(hash[:]) != location.Query().Get("code_challenge") {
		return errors.New("code verifier sent to the IdP does not match the code challenge from the login redirect")
	}

	return nil
}

func testOIDCLogin_CallbackOnAnotherReplica() (testName string, testFunc func(t *testing.T)) {
	return "OIDC login started on one replica can be completed on another", func(t *testing.T) {
		var condition string

		idp := helpers.NewMockIdentityProvider()
		defer idp.Close()

		cert, err := helpers.GenerateCertInfo()
		helpers.LogTestOutcome(t, "Should generate a management API certificate", err)

		replicaA, err := helpers.NewOIDCVerifier(idp.URL, cert)
		helpers.LogTestOutcome(t, "Should create the first management API replica", err)

		replicaB, err := helpers.NewOIDCVerifier(idp.URL, cert)
		helpers.LogTestOutcome(t, "Should create the second management API replica", err)

		var loginRes *http.Response
		{
			condition = "Should set a SameSite=Lax login_id cookie when starting the login flow"

			loginRes, err = helpers.StartOIDCLogin(replicaA, nil)
			if err == nil {
				if loginRes.StatusCode != http.StatusFound {
					err = fmt.Errorf("expected status 302, got %d", loginRes.StatusCode)
				} else if loginID := findCookie(loginRes.Cookies(), loginIDCookieName); loginID == nil {
					err = errors.New("login response did not set a login_id cookie")
				} else if loginID.SameSite != http.SameSiteLaxMode {
					err = fmt.Errorf("expected login_id cookie to be SameSite=Lax, got %v", loginID.SameSite)
				}
			}

			helpers.LogTestOutcome(t, condition, err)
		}

		var callbackRes *http.Response
		{
			condition = "Should decrypt the state and PKCE cookies on the other replica"

			callbackRes, err = helpers.FinishOIDCLogin(replicaB, loginRes)
			if err == nil {
				err = checkCodeExchange(idp, loginRes)
			}

			helpers.LogTestOutcome(t, condition, err)
		}

		{
			condition = "Should delete the login_id cookie on callback"

			loginID := findCookie(callbackRes.Cookies(), loginIDCookieName)
			if loginID == nil {
				err = errors.New("callback response did not reset the login_id cookie")
			} else if loginID.MaxAge >= 0 && !loginID.Expires.Before(time.Now()) {
				err = fmt.Errorf("expected login_id cookie to be expired, got %q", loginID.Raw)
			}

			helpers.LogTestOutcome(t, condition, err)
		}
	}
}

func testOIDCLogin_IgnoresStaleLoginID() (testName string, testFunc func(t *testing.T)) {
	return "OIDC login ignores a login_id cookie left over from an abandoned login", func(t *testing.T) {
		var condition string

		idp := helpers.NewMockIdentityProvider()
		defer idp.Close()

		cert, err := helpers.GenerateCertInfo()
		helpers.LogTestOutcome(t, "Should generate a management API certificate", err)

		verifier, err := helpers.NewOIDCVerifier(idp.URL, cert)
		helpers.LogTestOutcome(t, "Should create a management API replica", err)

		abandonedRes, err := helpers.StartOIDCLogin(verifier, nil)
		helpers.LogTestOutcome(t, "Should start a login flow that is then abandoned", err)

		staleLoginID := findCookie(abandonedRes.Cookies(), loginIDCookieName)
		if staleLoginID == nil {
			helpers.LogTestOutcome(t, "Should set a login_id cookie for the abandoned login", errors.New("no login_id cookie set"))
		}

		{
			condition = "Should complete a new login flow while the stale login_id cookie is still in the browser"

			loginRes, err := helpers.StartOIDCLogin(verifier, []*http.Cookie{staleLoginID})
			if err == nil {
				_, err = helpers.FinishOIDCLogin(verifier, loginRes)
			}

			if err == nil {
				err = checkCodeExchange(idp, loginRes)
			}

			helpers.LogTestOutcome(t, condition, err)
		}
	}
}

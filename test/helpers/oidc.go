package helpers

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"

	"github.com/canonical/lxd/shared"
	"github.com/canonical/microcloud-cluster-manager/internal/app/management-api/core/auth"
	"github.com/google/uuid"
)

// oidcTestHost is the management API host used for OIDC login requests in unit tests.
const oidcTestHost = "ma.lxd-cm.local"

// MockIdentityProvider is a minimal OIDC identity provider that serves the discovery document and records the PKCE
// code verifier sent on token exchange. The token exchange itself is always rejected, so a callback only reaches the
// IdP if it was able to decrypt the state and PKCE cookies set by the login handler.
type MockIdentityProvider struct {
	*httptest.Server

	mu           sync.Mutex
	codeVerifier string
}

// NewMockIdentityProvider starts a MockIdentityProvider. The caller must call Close when done.
func NewMockIdentityProvider() *MockIdentityProvider {
	idp := &MockIdentityProvider{}

	mux := http.NewServeMux()
	mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"issuer":                 idp.URL,
			"authorization_endpoint": idp.URL + "/authorize",
			"token_endpoint":         idp.URL + "/oauth/token",
			"jwks_uri":               idp.URL + "/keys",
		})
	})

	mux.HandleFunc("/oauth/token", func(w http.ResponseWriter, r *http.Request) {
		idp.mu.Lock()
		idp.codeVerifier = r.FormValue("code_verifier")
		idp.mu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"invalid_grant"}`))
	})

	idp.Server = httptest.NewServer(mux)

	return idp
}

// CodeVerifier returns the PKCE code verifier sent with the last token exchange request.
func (idp *MockIdentityProvider) CodeVerifier() string {
	idp.mu.Lock()
	defer idp.mu.Unlock()

	return idp.codeVerifier
}

// GenerateCertInfo generates a self-signed certificate for use as the management API certificate.
func GenerateCertInfo() (*shared.CertInfo, error) {
	certPEM, keyPEM, err := shared.GenerateMemCert(false, shared.CertOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to generate certificate: %w", err)
	}

	keyPair, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to load key pair: %w", err)
	}

	return shared.NewCertInfo(keyPair, nil, nil), nil
}

// NewOIDCVerifier returns a management API OIDC verifier for the given issuer. Verifiers created with the same
// certificate behave like replicas of the same management API deployment.
func NewOIDCVerifier(issuer string, cert *shared.CertInfo) (*auth.Verifier, error) {
	return auth.NewVerifier(issuer, "cluster-manager", "cluster-manager-secret", "cluster-manager", cert, nil)
}

// StartOIDCLogin calls the /oidc/login handler of the given verifier with the given request cookies and returns the
// response.
func StartOIDCLogin(verifier *auth.Verifier, cookies []*http.Cookie) (*http.Response, error) {
	state, err := auth.StateToken{RedirectURL: "/ui", ID: uuid.NewString()}.String()
	if err != nil {
		return nil, fmt.Errorf("failed to create state token: %w", err)
	}

	req := httptest.NewRequest(http.MethodGet, "https://"+oidcTestHost+"/oidc/login", nil)
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}

	rec := httptest.NewRecorder()
	verifier.Login(rec, req, state)

	return rec.Result(), nil
}

// FinishOIDCLogin follows the IdP redirect in loginRes back to the /oidc/callback handler of the given verifier,
// sending the cookies that were set by the login response, and returns the response.
func FinishOIDCLogin(verifier *auth.Verifier, loginRes *http.Response) (*http.Response, error) {
	location, err := loginRes.Location()
	if err != nil {
		return nil, fmt.Errorf("failed to get login redirect location: %w", err)
	}

	callbackURL := url.URL{
		Scheme:   "https",
		Host:     oidcTestHost,
		Path:     "/oidc/callback",
		RawQuery: url.Values{"state": {location.Query().Get("state")}, "code": {"authorization-code"}}.Encode(),
	}

	req := httptest.NewRequest(http.MethodGet, callbackURL.String(), nil)
	for _, cookie := range loginRes.Cookies() {
		req.AddCookie(cookie)
	}

	rec := httptest.NewRecorder()
	verifier.Callback(rec, req, "/ui")

	return rec.Result(), nil
}

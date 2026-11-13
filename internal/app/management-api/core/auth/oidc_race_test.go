package auth

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"
	"time"

	"github.com/canonical/lxd/shared"
)

type headerHookWriter struct {
	*httptest.ResponseRecorder
	once sync.Once
	hook func()
}

func (h *headerHookWriter) Header() http.Header {
	h.once.Do(h.hook)
	return h.ResponseRecorder.Header()
}

// TestLoginUsesRelyingPartySnapshot checks that a concurrent request with a different Host header cannot swap the
// relying party used by an in-flight login.
func TestLoginUsesRelyingPartySnapshot(t *testing.T) {
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"issuer":                 srv.URL,
			"authorization_endpoint": srv.URL + "/authorize",
			"token_endpoint":         srv.URL + "/token",
			"jwks_uri":               srv.URL + "/jwks",
			"userinfo_endpoint":      srv.URL + "/userinfo",
		})
	}))
	defer srv.Close()

	cert := shared.TestingKeyPair()
	o := &Verifier{
		issuer:               srv.URL,
		clientID:             "client",
		clientSecret:         "secret",
		audience:             "aud",
		clusterCert:          func() *shared.CertInfo { return cert },
		configExpiryInterval: time.Hour,
		httpClientFunc:       func() (*http.Client, error) { return srv.Client(), nil },
	}

	req := httptest.NewRequest(http.MethodGet, "https://legit.example/oidc/login", nil)
	req.Host = "legit.example"

	w := &headerHookWriter{ResponseRecorder: httptest.NewRecorder()}
	w.hook = func() {
		done := make(chan struct{})
		go func() {
			defer close(done)
			other := httptest.NewRequest(http.MethodGet, "https://other.example/oidc/login", nil)
			other.Host = "other.example"
			err := o.ensureConfig(other.Context(), other)
			if err != nil {
				t.Errorf("ensureConfig: %v", err)
			}
		}()
		<-done
	}

	o.Login(w, req, "state")

	loc := w.Result().Header.Get("Location")
	u, err := url.Parse(loc)
	if err != nil || loc == "" {
		t.Fatalf("no redirect location: %q (%v)", loc, err)
	}

	got := u.Query().Get("redirect_uri")
	if got != "https://legit.example/oidc/callback" {
		t.Fatalf("unexpected redirect_uri %q", got)
	}
}

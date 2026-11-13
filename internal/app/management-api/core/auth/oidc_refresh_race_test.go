package auth

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/canonical/lxd/shared"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/config"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/database"
	"github.com/go-jose/go-jose/v4"
	"github.com/google/uuid"
	"github.com/zitadel/oidc/v3/pkg/oidc"
)

type refreshRaceEnv struct {
	verifier     *Verifier
	host         string
	mkIDToken    func(exp time.Time, iat time.Time) string
	refreshCalls func() int
}

// newRefreshRaceEnv sets up a verifier backed by the local development Postgres and a mock IdP that rotates
// refresh tokens and rejects reuse. The test is skipped if Postgres is not reachable.
func newRefreshRaceEnv(t *testing.T) *refreshRaceEnv {
	t.Helper()
	const clientID = "client"

	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}

	signer, err := jose.NewSigner(jose.SigningKey{Algorithm: jose.RS256, Key: jose.JSONWebKey{Key: priv, KeyID: "k1"}}, (&jose.SignerOptions{}).WithType("JWT"))
	if err != nil {
		t.Fatal(err)
	}

	subject := "refresh-race-" + uuid.NewString()
	var issuer string
	mkIDToken := func(exp time.Time, iat time.Time) string {
		payload, _ := json.Marshal(map[string]any{
			"iss": issuer, "sub": subject, "aud": []string{clientID},
			"exp": exp.Unix(), "iat": iat.Unix(), "email": "u@example.com", "name": "u",
		})

		jws, err := signer.Sign(payload)
		if err != nil {
			t.Fatal(err)
		}

		s, _ := jws.CompactSerialize()
		return s
	}

	var mu sync.Mutex
	used := map[string]bool{}
	valid := map[string]bool{"rt-0": true}
	counter := 0
	refreshCalls := 0

	mux := http.NewServeMux()
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	issuer = srv.URL

	mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"issuer":                                issuer,
			"authorization_endpoint":                issuer + "/authorize",
			"token_endpoint":                        issuer + "/token",
			"jwks_uri":                              issuer + "/jwks",
			"userinfo_endpoint":                     issuer + "/userinfo",
			"id_token_signing_alg_values_supported": []string{"RS256"},
			"response_types_supported":              []string{"code"},
			"subject_types_supported":               []string{"public"},
			"token_endpoint_auth_methods_supported": []string{"client_secret_basic", "client_secret_post"},
		})
	})
	mux.HandleFunc("/jwks", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(jose.JSONWebKeySet{Keys: []jose.JSONWebKey{{Key: &priv.PublicKey, KeyID: "k1", Algorithm: "RS256", Use: "sig"}}})
	})
	mux.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		rt := r.Form.Get("refresh_token")

		// Give concurrent requests a chance to overlap.
		time.Sleep(200 * time.Millisecond)

		mu.Lock()
		defer mu.Unlock()
		refreshCalls++
		if used[rt] || !valid[rt] {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"error":"invalid_grant","error_description":"refresh token reuse detected"}`))
			return
		}

		used[rt] = true
		counter++
		newRT := fmt.Sprintf("rt-%d", counter)
		valid[newRT] = true
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token":  fmt.Sprintf("at-%d", counter),
			"token_type":    "Bearer",
			"refresh_token": newRT,
			"expires_in":    3600,
			"id_token":      mkIDToken(time.Now().Add(time.Hour), time.Now()),
		})
	})

	cert := shared.TestingKeyPair()

	db, err := database.NewDB(database.DBConfig{DBHost: "localhost:5432", DBUser: "admin", DBPassword: "admin", DBName: "cm", DBDisableTLS: true, DBMaxIdleConns: 5, DBMaxOpenConns: 20})
	if err != nil {
		t.Skipf("Postgres not available: %v", err)
	}

	v := &Verifier{
		issuer:               issuer,
		clientID:             clientID,
		clientSecret:         "secret",
		clusterCert:          func() *shared.CertInfo { return cert },
		configExpiryInterval: time.Hour,
		httpClientFunc:       func() (*http.Client, error) { return srv.Client(), nil },
		sessionHandler:       NewSessionHandler(db),
	}

	return &refreshRaceEnv{
		verifier:  v,
		host:      "ma.example",
		mkIDToken: mkIDToken,
		refreshCalls: func() int {
			mu.Lock()
			defer mu.Unlock()
			return refreshCalls
		},
	}
}

// startSessionWithExpiredIDToken creates a session whose ID token has expired and returns cookies for it with the
// given session JWT expiry.
func (e *refreshRaceEnv) startSessionWithExpiredIDToken(t *testing.T, jwtExpiry time.Time) (uuid.UUID, []*http.Cookie) {
	t.Helper()

	baseReq := httptest.NewRequest(http.MethodGet, "https://"+e.host+"/1.0/clusters", nil)
	err := e.verifier.ensureConfig(context.Background(), baseReq)
	if err != nil {
		t.Fatalf("ensureConfig: %v", err)
	}

	userSecret, err := config.CreateUserSecret()
	if err != nil {
		t.Fatal(err)
	}

	expiredID := e.mkIDToken(time.Now().Add(-time.Hour), time.Now().Add(-2*time.Hour))
	enc, err := EncryptTokens(expiredID, "at-0", "rt-0", userSecret)
	if err != nil {
		t.Fatal(err)
	}

	claims := &oidc.IDTokenClaims{}
	claims.Subject = "refresh-race"
	claims.Email = "u@example.com"
	claims.Name = "u"
	sessionID, _, err := e.verifier.sessionHandler.StartSession(baseReq, claims, enc)
	if err != nil {
		t.Skipf("Postgres not available: %v", err)
	}

	rec := httptest.NewRecorder()
	err = e.verifier.setCookies(rec, *sessionID, userSecret, jwtExpiry)
	if err != nil {
		t.Fatal(err)
	}

	return *sessionID, rec.Result().Cookies()
}

func (e *refreshRaceEnv) authConcurrently(cookies []*http.Cookie, n int) ([]error, []*httptest.ResponseRecorder) {
	errs := make([]error, n)
	recs := make([]*httptest.ResponseRecorder, n)
	var wg sync.WaitGroup
	for i := range n {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			req := httptest.NewRequest(http.MethodGet, "https://"+e.host+"/1.0/clusters", nil)
			for _, c := range cookies {
				req.AddCookie(c)
			}

			recs[i] = httptest.NewRecorder()
			errs[i] = e.verifier.Auth(req.Context(), recs[i], req)
		}(i)
	}

	wg.Wait()
	return errs, recs
}

// TestConcurrentExpiredSessionRenewalRedeemsRefreshTokenOnce checks that parallel requests carrying the same expired
// session cookie redeem the refresh token only once and all end up on the same renewed session.
func TestConcurrentExpiredSessionRenewalRedeemsRefreshTokenOnce(t *testing.T) {
	env := newRefreshRaceEnv(t)
	sessionID, cookies := env.startSessionWithExpiredIDToken(t, time.Now().Add(-time.Minute))

	errs, recs := env.authConcurrently(cookies, 3)
	for i, err := range errs {
		if err != nil {
			t.Fatalf("request %d failed: %v", i, err)
		}
	}

	if calls := env.refreshCalls(); calls != 1 {
		t.Fatalf("Expected exactly 1 refresh call, got %d", calls)
	}

	newSession := ""
	for i, rec := range recs {
		for _, c := range rec.Result().Cookies() {
			if c.Name != cookieNameSessionToken {
				continue
			}

			if newSession == "" {
				newSession = c.Value
			}

			if c.Value != newSession {
				t.Fatalf("request %d received a different session than the others", i)
			}
		}
	}

	if newSession == "" {
		t.Fatal("No renewed session cookie was set")
	}

	_, err := env.verifier.sessionHandler.GetSessionByID(context.Background(), sessionID)
	if err == nil {
		t.Fatal("Expected the expired session to be deleted")
	}
}

// TestConcurrentIDTokenRefreshRedeemsRefreshTokenOnce checks that parallel requests with a valid session whose ID
// token has expired redeem the refresh token only once.
func TestConcurrentIDTokenRefreshRedeemsRefreshTokenOnce(t *testing.T) {
	env := newRefreshRaceEnv(t)
	sessionID, cookies := env.startSessionWithExpiredIDToken(t, time.Now().Add(time.Hour))

	errs, _ := env.authConcurrently(cookies, 3)
	for i, err := range errs {
		if err != nil {
			t.Fatalf("request %d failed: %v", i, err)
		}
	}

	if calls := env.refreshCalls(); calls != 1 {
		t.Fatalf("Expected exactly 1 refresh call, got %d", calls)
	}

	_, err := env.verifier.sessionHandler.GetSessionByID(context.Background(), sessionID)
	if err != nil {
		t.Fatalf("Expected the session to still exist: %v", err)
	}
}

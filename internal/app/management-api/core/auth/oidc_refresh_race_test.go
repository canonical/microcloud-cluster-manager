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
	"github.com/canonical/lxd/shared/api"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/config"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/database"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/database/store"
	"github.com/go-jose/go-jose/v4"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/zitadel/oidc/v3/pkg/oidc"
)

type refreshRaceEnv struct {
	verifier     *Verifier
	host         string
	mkIDToken    func(exp time.Time, iat time.Time) string
	refreshCalls func() int
	// subject is the subject of ID tokens issued by the mock IdP.
	subject string
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
		subject:   subject,
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
	sessionID, userSecret := e.startSessionWithExpiredIDTokenAndSecret(t)
	return sessionID, e.cookies(t, sessionID, userSecret, jwtExpiry)
}

// cookies returns session cookies for the given session and user secret with the given session JWT expiry.
func (e *refreshRaceEnv) cookies(t *testing.T, sessionID uuid.UUID, userSecret string, jwtExpiry time.Time) []*http.Cookie {
	t.Helper()
	rec := httptest.NewRecorder()
	err := e.verifier.setCookies(rec, sessionID, userSecret, jwtExpiry)
	if err != nil {
		t.Fatal(err)
	}

	return rec.Result().Cookies()
}

// startSessionWithExpiredIDTokenAndSecret creates a session whose ID token has expired and returns its ID and user
// secret.
func (e *refreshRaceEnv) startSessionWithExpiredIDTokenAndSecret(t *testing.T) (uuid.UUID, string) {
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

	return *sessionID, userSecret
}

// auth authenticates a request with the given cookies and returns the session credentials provider it set up.
func (e *refreshRaceEnv) auth(cookies []*http.Cookie) (*SessionCredentialsProvider, *httptest.ResponseRecorder, error) {
	req := httptest.NewRequest(http.MethodGet, "https://"+e.host+"/1.0/clusters", nil)
	for _, c := range cookies {
		req.AddCookie(c)
	}

	rec := httptest.NewRecorder()
	err := e.verifier.Auth(req.Context(), rec, req)
	if err != nil {
		return nil, rec, err
	}

	provider, _ := req.Context().Value(SessionCredentialsProviderKey).(*SessionCredentialsProvider)
	return provider, rec, nil
}

func (e *refreshRaceEnv) authConcurrently(cookies []*http.Cookie, n int) ([]error, []*httptest.ResponseRecorder) {
	errs := make([]error, n)
	recs := make([]*httptest.ResponseRecorder, n)
	var wg sync.WaitGroup
	for i := range n {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_, recs[i], errs[i] = e.auth(cookies)
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

// TestValidSessionFollowsRenewalOfSameSession checks that a request that resolved a session before it was renewed
// by another request, through a still-valid session JWT, keeps working afterwards instead of failing on the deleted
// session.
func TestValidSessionFollowsRenewalOfSameSession(t *testing.T) {
	env := newRefreshRaceEnv(t)
	sessionID, userSecret := env.startSessionWithExpiredIDTokenAndSecret(t)
	validCookies := env.cookies(t, sessionID, userSecret, time.Now().Add(time.Hour))
	expiredCookies := env.cookies(t, sessionID, userSecret, time.Now().Add(-time.Minute))

	// A request with a valid session JWT refreshes the expired ID token of the session.
	provider, _, err := env.auth(validCookies)
	if err != nil {
		t.Fatalf("Auth with valid session failed: %v", err)
	}

	// A request carrying the expired session JWT renews, and so deletes, the same session.
	_, _, err = env.auth(expiredCookies)
	if err != nil {
		t.Fatalf("Auth with expired session failed: %v", err)
	}

	_, err = env.verifier.sessionHandler.GetSessionByID(context.Background(), sessionID)
	if err == nil {
		t.Fatal("Expected the expired session to be deleted")
	}

	// The provider resolved before the renewal must still be able to refresh tokens.
	_, err = provider.RefreshSessionTokens(context.Background())
	if err != nil {
		t.Fatalf("Refreshing tokens after renewal failed: %v", err)
	}

	if provider.sessionID == sessionID {
		t.Fatal("Expected the provider to follow the renewed session")
	}

	_, err = env.verifier.sessionHandler.GetSessionByID(context.Background(), provider.sessionID)
	if err != nil {
		t.Fatalf("Expected the renewed session to exist: %v", err)
	}

	// A later request with the still-valid JWT of the deleted session follows the renewal instead of logging out.
	followProvider, rec, err := env.auth(validCookies)
	if err != nil {
		t.Fatalf("Auth with valid JWT of renewed session failed: %v", err)
	}

	if followProvider.sessionID != provider.sessionID {
		t.Fatal("Expected the request to use the renewed session")
	}

	for _, c := range rec.Result().Cookies() {
		if c.Name == cookieNameSessionToken && c.Value == "" {
			t.Fatal("Expected session cookies not to be deleted")
		}
	}

	if calls := env.refreshCalls(); calls != 2 {
		t.Fatalf("Expected exactly 2 refresh calls, got %d", calls)
	}
}

// TestConcurrentRefreshAndRenewalOfSameSession checks that concurrent requests on the same session, some with a valid
// and some with an expired session JWT, redeem the refresh token only once and all end up with usable credentials.
func TestConcurrentRefreshAndRenewalOfSameSession(t *testing.T) {
	env := newRefreshRaceEnv(t)
	sessionID, userSecret := env.startSessionWithExpiredIDTokenAndSecret(t)
	validCookies := env.cookies(t, sessionID, userSecret, time.Now().Add(time.Hour))
	expiredCookies := env.cookies(t, sessionID, userSecret, time.Now().Add(-time.Minute))

	const n = 6
	providers := make([]*SessionCredentialsProvider, n)
	errs := make([]error, n)
	var wg sync.WaitGroup
	for i := range n {
		cookies := validCookies
		if i%2 == 1 {
			cookies = expiredCookies
		}

		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			providers[i], _, errs[i] = env.auth(cookies)
		}(i)
	}

	wg.Wait()
	for i, err := range errs {
		if err != nil {
			t.Fatalf("request %d failed: %v", i, err)
		}
	}

	if calls := env.refreshCalls(); calls != 1 {
		t.Fatalf("Expected exactly 1 refresh call, got %d", calls)
	}

	for i, provider := range providers {
		_, err := provider.RefreshSessionTokens(context.Background())
		if err != nil {
			t.Fatalf("request %d could not refresh tokens: %v", i, err)
		}
	}
}

// logout logs out a request carrying the given cookies.
func (e *refreshRaceEnv) logout(cookies []*http.Cookie) {
	req := httptest.NewRequest(http.MethodGet, "https://"+e.host+"/oidc/logout", nil)
	for _, c := range cookies {
		req.AddCookie(c)
	}

	e.verifier.Logout(httptest.NewRecorder(), req)
}

// identityHasSession reports whether any session is left for the identity the test session was started with, or for
// the identity of ID tokens issued by the mock IdP, which renewed sessions belong to.
func (e *refreshRaceEnv) identityHasSession(t *testing.T) bool {
	t.Helper()
	var exists bool
	err := e.verifier.sessionHandler.db.Transaction(context.Background(), func(ctx context.Context, tx *sqlx.Tx) error {
		for _, subject := range []string{"refresh-race", e.subject} {
			identity, err := store.GetIdentity(ctx, tx, subject)
			if api.StatusErrorCheck(err, http.StatusNotFound) {
				continue
			}

			if err != nil {
				return err
			}

			found, err := store.SessionExistsForIdentity(ctx, tx, identity.ID)
			if err != nil {
				return err
			}

			exists = exists || found
		}

		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	return exists
}

// TestLogoutWithStaleCookieRevokesRenewedSession checks that logging out with the cookie of a session that has just
// been renewed also revokes the renewed session.
func TestLogoutWithStaleCookieRevokesRenewedSession(t *testing.T) {
	env := newRefreshRaceEnv(t)
	sessionID, userSecret := env.startSessionWithExpiredIDTokenAndSecret(t)
	validCookies := env.cookies(t, sessionID, userSecret, time.Now().Add(time.Hour))
	expiredCookies := env.cookies(t, sessionID, userSecret, time.Now().Add(-time.Minute))

	provider, _, err := env.auth(expiredCookies)
	if err != nil {
		t.Fatalf("Auth with expired session failed: %v", err)
	}

	if provider.sessionID == sessionID {
		t.Fatal("Expected the session to be renewed")
	}

	env.logout(expiredCookies)

	if env.identityHasSession(t) {
		t.Fatal("Expected logout to revoke the renewed session")
	}

	_, _, err = env.auth(expiredCookies)
	if err == nil {
		t.Fatal("Expected the stale expired cookie to be rejected after logout")
	}

	_, _, err = env.auth(validCookies)
	if err == nil {
		t.Fatal("Expected the stale valid cookie to be rejected after logout")
	}
}

// TestLogoutConcurrentWithRenewalRevokesAllSessions checks that a logout racing with renewals of the same session
// leaves no session behind.
func TestLogoutConcurrentWithRenewalRevokesAllSessions(t *testing.T) {
	env := newRefreshRaceEnv(t)
	_, cookies := env.startSessionWithExpiredIDToken(t, time.Now().Add(-time.Minute))

	var wg sync.WaitGroup
	for range 3 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _, _ = env.auth(cookies)
		}()
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		env.logout(cookies)
	}()

	wg.Wait()

	if env.identityHasSession(t) {
		t.Fatal("Expected logout to leave no session behind")
	}
}

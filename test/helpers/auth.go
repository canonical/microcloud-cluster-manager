package helpers

import (
	"context"
	"crypto/sha512"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	"github.com/canonical/lxd/shared"
	"github.com/canonical/lxd/shared/api"
	"github.com/canonical/microcloud-cluster-manager/internal/app/management-api/core/auth"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/types"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/hkdf"
)

func LoginToManagementAPI(e *Environment, username string, password string, serverCert *x509.Certificate) ([]*http.Cookie, error) {
	jar, _ := cookiejar.New(nil)

	// Add the public key to the CA pool to make it trusted.
	tlsConfig := shared.InitTLSConfig()
	rootCAs, err := x509.SystemCertPool()
	if err != nil || rootCAs == nil {
		tlsConfig.RootCAs = x509.NewCertPool()
	} else {
		tlsConfig.RootCAs = rootCAs
	}
	serverCert.IsCA = true
	serverCert.KeyUsage = x509.KeyUsageCertSign
	tlsConfig.RootCAs.AddCert(serverCert)

	client := &http.Client{
		Jar: jar,
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}

	// Request app login page
	resp, err := client.Get("https://" + e.ManagementAPIHostPort() + "/oidc/login")
	if err != nil {
		return nil, err
	}
	_, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	err = resp.Body.Close()
	if err != nil {
		return nil, err
	}

	// Capture redirect to IdP
	idpURL := resp.Request.URL

	// Submit login form to IdP
	form := url.Values{}
	form.Set("username", username)
	form.Set("password", password)

	// Forward query params (state, nonce, redirect_uri, etc.)
	q := idpURL.Query()
	for k, v := range q {
		form.Set(k, v[0])
	}

	loginAction := idpURL.Scheme + "://" + idpURL.Host + idpURL.Path
	if loginAction != "https://dev-h6c02msuggpi6ijh.eu.auth0.com/u/login" {
		return nil, fmt.Errorf("disallowed login action: %q", loginAction)
	}

	resp, err = client.Post(
		loginAction,
		"application/x-www-form-urlencoded",
		strings.NewReader(form.Encode()),
	)
	if err != nil {
		return nil, err
	}
	_, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	err = resp.Body.Close()
	if err != nil {
		return nil, err
	}

	// get stored cookies
	cookies := jar.Cookies(&url.URL{Scheme: "https", Host: e.managementAPIHost})

	return cookies, nil
}

func GetManagementAPIAuthorizor() (*auth.ManagementAPIAuthorizor, error) {
	return auth.NewManagementAPIAuthorizor()
}

func GetContextWithUserInfo(groups []string) context.Context {
	userInfo := &types.UserInfo{
		Groups: groups,
	}
	ctx := context.WithValue(context.Background(), types.UserInfoKey, userInfo)
	return ctx
}

// GetCookies performs a full login and returns the resulting session cookies.
func GetCookies(env *Environment, username string, password string) ([]*http.Cookie, error) {
	certPublicKey, err := env.ManagementAPICert().PublicKeyX509()
	if err != nil {
		return nil, fmt.Errorf("failed to get management API cert: %w", err)
	}

	cookies, err := LoginToManagementAPI(env, username, password, certPublicKey)
	if err != nil {
		return nil, fmt.Errorf("login failed: %w", err)
	}

	return cookies, nil
}

// LogoutFromManagementAPI calls the /oidc/logout endpoint with the given session cookies,
// deleting the session from the server. Redirects are not followed.
func LogoutFromManagementAPI(env *Environment, cookies []*http.Cookie) error {
	certPublicKey, err := env.ManagementAPICert().PublicKeyX509()
	if err != nil {
		return fmt.Errorf("failed to get management API cert: %w", err)
	}

	tlsClient, err := NewTLSHTTPClient(api.URL{}, nil, certPublicKey, env.ManagementAPIHost())
	if err != nil {
		return fmt.Errorf("failed to create TLS client: %w", err)
	}

	tlsClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	logoutURL := fmt.Sprintf("https://%s/oidc/logout", env.ManagementAPIHostPort())

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, logoutURL, nil)
	if err != nil {
		return fmt.Errorf("failed to build logout request: %w", err)
	}

	for _, c := range cookies {
		req.AddCookie(c)
	}

	res, err := tlsClient.Do(req)
	if err != nil && !errors.Is(err, http.ErrUseLastResponse) {
		return fmt.Errorf("logout request failed: %w", err)
	}

	defer res.Body.Close()
	_, _ = io.Copy(io.Discard, res.Body)
	return nil
}

func AddCookiesToRequest(cookies []*http.Cookie) func(*http.Request) error {
	return func(r *http.Request) error {
		for _, cookie := range cookies {
			r.AddCookie(cookie)
		}
		return nil
	}
}

// DeriveSessionSigningKey derives the signing key for a session token using HKDF-SHA512.
// It uses the same derivation method as the server:
// HKDF-SHA512(privateKey, sessionIDBytes, "SESSION_ID_TOKEN_INTEGRITY", 64).
func DeriveSessionSigningKey(sessionID uuid.UUID, privateKeyBytes []byte) ([]byte, error) {
	salt, err := sessionID.MarshalBinary()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal session ID: %w", err)
	}

	prk := hkdf.Extract(sha512.New, privateKeyBytes, salt)
	kdf := hkdf.Expand(sha512.New, prk, []byte("SESSION_ID_TOKEN_INTEGRITY"))
	signingKey := make([]byte, 64)
	if _, err = io.ReadFull(kdf, signingKey); err != nil {
		return nil, fmt.Errorf("failed to derive session token signing key: %w", err)
	}

	return signingKey, nil
}

// CreateExpiredSessionToken creates a session token that is expired but bears a correct signature.
// The token is signed with the provided key and has an expiry one hour in the past.
func CreateExpiredSessionToken(sessionID uuid.UUID, signingKey []byte) (string, error) {
	expiredClaims := &jwt.RegisteredClaims{
		Subject:   sessionID.String(),
		IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
	}
	expiredToken := jwt.NewWithClaims(jwt.SigningMethodHS512, expiredClaims)
	expiredTokenSigned, err := expiredToken.SignedString(signingKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign expired session JWT: %w", err)
	}

	return expiredTokenSigned, nil
}

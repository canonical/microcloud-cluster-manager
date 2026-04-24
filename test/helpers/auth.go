package helpers

import (
	"context"
	"crypto/x509"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"

	"github.com/canonical/lxd/shared"
	"github.com/canonical/lxd/shared/api"
	"github.com/canonical/microcloud-cluster-manager/internal/app/management-api/core/auth"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/types"
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

// DoRequestWithCookies makes a request to an endpoint with the provided cookies,
// and returns the HTTP status code.
func DoRequestWithCookies(env *Environment, path string, method string, cookies []*http.Cookie) (int, error) {
	certPublicKey, err := env.ManagementAPICert().PublicKeyX509()
	if err != nil {
		return 0, fmt.Errorf("failed to get management API cert: %w", err)
	}

	tlsClient, err := NewTLSHTTPClient(api.URL{}, nil, certPublicKey, env.ManagementAPIHost())
	if err != nil {
		return 0, fmt.Errorf("failed to create TLS client: %w", err)
	}

	fullPath := fmt.Sprintf("https://%s%s", env.ManagementAPIHostPort(), path)
	req, err := http.NewRequestWithContext(context.Background(), method, fullPath, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to build request: %w", err)
	}

	for _, c := range cookies {
		req.AddCookie(c)
	}

	resp, err := tlsClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("request failed: %w", err)
	}

	defer resp.Body.Close()
	return resp.StatusCode, nil
}

// GetValidCookies performs a full login and returns the resulting session cookies.
func GetValidCookies(env *Environment, username string, password string) ([]*http.Cookie, error) {
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

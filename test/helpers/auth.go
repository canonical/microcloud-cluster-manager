package helpers

import (
	"crypto/tls"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
)

func LoginToManagementAPI(e *Environment, username string, password string) ([]*http.Cookie, error) {
	jar, _ := cookiejar.New(nil)

	client := &http.Client{
		Jar: jar,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, // ignore cert errors
			},
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

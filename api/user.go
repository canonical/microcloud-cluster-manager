package api

import (
	"net/http"

	"github.com/canonical/lxd/lxd/response"

	"github.com/canonical/microcluster/rest"
	"github.com/canonical/microcluster/state"

	"github.com/canonical/lxd-site-manager/auth"
)

var oidcLoginCmd = rest.Endpoint{
	Path: "oidc/login",
	Get:  rest.EndpointAction{Handler: oidcLogin, AllowUntrusted: true},
}

var oidcCallbackCmd = rest.Endpoint{
	Path: "oidc/callback",
	Get:  rest.EndpointAction{Handler: oidcCallback, AllowUntrusted: true},
}

var oidcLogoutCmd = rest.Endpoint{
	Path: "oidc/logout",
	Get:  rest.EndpointAction{Handler: oidcLogout, AllowUntrusted: true},
}

func oidcLogin(s *state.State, r *http.Request) response.Response {
	// Setup OIDC authentication.
	// TODO: not ideal to do this here, should be done in a more central place.
	// TODO: how does the proxy work?
	oidcVerifier, err := auth.GetOIDCVerifierFromContext(s.Context)
	if err != nil {
		return response.SmartError(err)
	}

	loginHandler := func(w http.ResponseWriter) error {
		oidcVerifier.Login(w, r)
		return nil
	}

	return response.ManualResponse(loginHandler)
}

func oidcCallback(s *state.State, r *http.Request) response.Response {
	oidcVerifier, err := auth.GetOIDCVerifierFromContext(s.Context)
	if err != nil {
		return response.SmartError(err)
	}

	callbackHandler := func(w http.ResponseWriter) error {
		oidcVerifier.Callback(w, r)
		return nil
	}

	return response.ManualResponse(callbackHandler)
}

func oidcLogout(s *state.State, r *http.Request) response.Response {
	oidcVerifier, err := auth.GetOIDCVerifierFromContext(s.Context)
	if err != nil {
		return response.SmartError(err)
	}

	logoutHandler := func(w http.ResponseWriter) error {
		oidcVerifier.Logout(w, r)
		return nil
	}

	return response.ManualResponse(logoutHandler)
}

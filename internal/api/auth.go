package api

import (
	"net/http"

	"github.com/canonical/lxd/lxd/response"
	"github.com/canonical/microcluster/rest"
	microState "github.com/canonical/microcluster/state"
	"github.com/google/uuid"

	"github.com/canonical/lxd-cluster-manager/internal/api/types"
	"github.com/canonical/lxd-cluster-manager/internal/oidc"
	"github.com/canonical/lxd-cluster-manager/internal/state"
)

var oidcLoginCmd = func(s *state.ClusterManagerState) rest.Endpoint {
	return rest.Endpoint{
		Path: "oidc/login",
		Get:  rest.EndpointAction{Handler: oidcLogin(s), AllowUntrusted: true},
	}
}

var oidcCallbackCmd = func(s *state.ClusterManagerState) rest.Endpoint {
	return rest.Endpoint{
		Path: "oidc/callback",
		Get:  rest.EndpointAction{Handler: oidcCallback(s), AllowUntrusted: true},
	}
}

var oidcLogoutCmd = func(s *state.ClusterManagerState) rest.Endpoint {
	return rest.Endpoint{
		Path: "oidc/logout",
		Get:  rest.EndpointAction{Handler: oidcLogout(s), AllowUntrusted: true},
	}
}

func oidcLogin(s *state.ClusterManagerState) types.EndpointHandler {
	return func(innerState microState.State, r *http.Request) response.Response {
		redirectURL := r.URL.Query().Get("next")

		stateToken := oidc.StateToken{
			RedirectURL: redirectURL,
			ID:          uuid.New().String(),
		}

		state, err := stateToken.String()
		if err != nil {
			return response.InternalError(err)
		}

		loginHandler := func(w http.ResponseWriter) error {
			s.OIDCVerifier.Login(w, r, state)
			return nil
		}

		return response.ManualResponse(loginHandler)
	}
}

func oidcCallback(s *state.ClusterManagerState) types.EndpointHandler {
	return func(innerState microState.State, r *http.Request) response.Response {
		state := r.URL.Query().Get("state")
		stateToken, err := oidc.DecodeStateToken(state)
		if err != nil {
			return response.InternalError(err)
		}

		callbackHandler := func(w http.ResponseWriter) error {
			s.OIDCVerifier.Callback(w, r, stateToken.RedirectURL)
			return nil
		}

		return response.ManualResponse(callbackHandler)
	}
}

func oidcLogout(s *state.ClusterManagerState) types.EndpointHandler {
	return func(innerState microState.State, r *http.Request) response.Response {
		redirectURL := r.URL.Query().Get("next")

		logoutHandler := func(w http.ResponseWriter) error {
			s.OIDCVerifier.Logout(w, r, redirectURL)
			return nil
		}

		return response.ManualResponse(logoutHandler)
	}
}

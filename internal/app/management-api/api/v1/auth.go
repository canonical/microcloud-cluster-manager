package v1

import (
	"fmt"
	"net/http"

	"github.com/canonical/lxd/lxd/response"
	"github.com/canonical/microcloud-cluster-manager/internal/app/management-api/core/auth"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/logger"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/types"
	"github.com/google/uuid"
)

// Auth is the OIDC authentication endpoint group.
var Auth = types.RouteGroup{
	IsRoot: true,
	Prefix: "oidc",
	Endpoints: []types.Endpoint{
		{
			Path:    "login",
			Method:  http.MethodGet,
			Handler: login,
		},
		{
			Path:    "callback",
			Method:  http.MethodGet,
			Handler: callback,
		},
		{
			Path:    "logout",
			Method:  http.MethodGet,
			Handler: logout,
		},
	},
}

func login(rc types.RouteConfig) types.EndpointHandler {
	verifier, ok := rc.Auth.(*auth.Verifier)
	return func(w http.ResponseWriter, r *http.Request) error {
		if !ok {
			logger.Log.Info("AUTHN oidc authenticator missing")
			return response.InternalError(fmt.Errorf("oidc authenticator missing")).Render(w, r)
		}

		stateToken := auth.StateToken{
			RedirectURL: "/ui",
			ID:          uuid.New().String(),
		}

		state, err := stateToken.String()
		if err != nil {
			logger.Log.Info("AUTHN failed to create OIDC state token")
			return response.InternalError(err).Render(w, r)
		}

		loginHandler := func(w http.ResponseWriter) error {
			verifier.Login(w, r, state)
			return nil
		}

		return response.ManualResponse(loginHandler).Render(w, r)
	}
}

func callback(rc types.RouteConfig) types.EndpointHandler {
	verifier, ok := rc.Auth.(*auth.Verifier)
	return func(w http.ResponseWriter, r *http.Request) error {
		if !ok {
			logger.Log.Info("AUTHN oidc authenticator missing")
			return response.InternalError(fmt.Errorf("oidc authenticator missing")).Render(w, r)
		}

		state := r.URL.Query().Get("state")
		stateToken, err := auth.DecodeStateToken(state)
		if err != nil {
			logger.Log.Info("AUTHN invalid OIDC state token")
			return response.InternalError(err).Render(w, r)
		}

		callbackHandler := func(w http.ResponseWriter) error {
			verifier.Callback(w, r, stateToken.RedirectURL)
			return nil
		}

		return response.ManualResponse(callbackHandler).Render(w, r)
	}
}

func logout(rc types.RouteConfig) types.EndpointHandler {
	verifier, ok := rc.Auth.(*auth.Verifier)
	return func(w http.ResponseWriter, r *http.Request) error {
		if !ok {
			return response.InternalError(fmt.Errorf("oidc authenticator missing")).Render(w, r)
		}

		logoutHandler := func(w http.ResponseWriter) error {
			verifier.Logout(w, r)
			return nil
		}

		return response.ManualResponse(logoutHandler).Render(w, r)
	}
}

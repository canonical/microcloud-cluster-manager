package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/canonical/lxd/lxd/response"
	"github.com/canonical/microcluster/rest"
	microState "github.com/canonical/microcluster/state"
	"github.com/google/uuid"

	"github.com/canonical/lxd-site-manager/internal/api/types"
	"github.com/canonical/lxd-site-manager/internal/client"
	"github.com/canonical/lxd-site-manager/internal/oidc"
	"github.com/canonical/lxd-site-manager/internal/state"
)

var oidcLoginCmd = func(s *state.SiteManagerState) rest.Endpoint {
	return rest.Endpoint{
		Path: "oidc/login",
		Get:  rest.EndpointAction{Handler: oidcLogin(s), AllowUntrusted: true},
	}
}

var oidcCallbackCmd = func(s *state.SiteManagerState) rest.Endpoint {
	return rest.Endpoint{
		Path: "oidc/callback",
		Get:  rest.EndpointAction{Handler: oidcCallback(s), AllowUntrusted: true},
		// FIXME: microcluster does not allow adjusting requets headers when forwarding requests using built-in client functions
		// This is the reason why we need this POST /oidc/callback endpoint so that we can forward cookies in the request payload
		// We should remove this endpoint once the microcluster is updated to allow adjusting request headers
		Post: rest.EndpointAction{Handler: oidcCallback(s), AllowUntrusted: true},
	}
}

var oidcLogoutCmd = func(s *state.SiteManagerState) rest.Endpoint {
	return rest.Endpoint{
		Path: "oidc/logout",
		Get:  rest.EndpointAction{Handler: oidcLogout(s), AllowUntrusted: true},
	}
}

func oidcLogin(s *state.SiteManagerState) types.EndpointHandler {
	return func(innerState microState.State, r *http.Request) response.Response {
		redirectURL := r.URL.Query().Get("next")

		stateToken := oidc.StateToken{
			MemberAddress: innerState.Address().URL.String(),
			RedirectURL:   redirectURL,
			ID:            uuid.New().String(),
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

func oidcCallback(s *state.SiteManagerState) types.EndpointHandler {
	return func(innerState microState.State, r *http.Request) response.Response {
		// FIXME: We have to set the forwarded query parameter for a forwarded request because microcluster does not allow adjusting request headers
		// We should remove this once the microcluster is updated to allow adjusting request headers
		forwarded := r.URL.Query().Get("forwarded")
		if forwarded != "" {
			return handleForwardedCallback(s, r)
		}

		return handleCallback(s, innerState, r)
	}
}

func handleCallback(siteManagerState *state.SiteManagerState, microState microState.State, r *http.Request) response.Response {
	state := r.URL.Query().Get("state")
	stateToken, err := oidc.DecodeStateToken(state)
	if err != nil {
		return response.InternalError(err)
	}

	// If the member address in the state token is different from the current member address, then forward the request to the correct member
	currentMemberAddress := microState.Address().URL.String()
	if stateToken.MemberAddress != currentMemberAddress {
		url, err := url.Parse(stateToken.MemberAddress)
		if err != nil {
			return response.SmartError(fmt.Errorf("Failed to parse member address %q: %w", stateToken.MemberAddress, err))
		}

		targetClient, err := siteManagerState.MicroCluster.RemoteClient(url.Host)
		if err != nil {
			return response.SmartError(fmt.Errorf("Failed to get a client for cluster member with address %q: %w", stateToken.MemberAddress, err))
		}

		// Forward the request to the target client
		// FIXME: currently microcluser does not return a response from the client function
		// Once this is fixed, we should return the response from the client function since cookies would be set already in the response
		// Then we could just copy over the cookies from the response to the current response
		tokens, err := client.CallbackForwardCmd(r.Context(), targetClient, r)
		if err != nil {
			targetClientURL := targetClient.URL()
			return response.SmartError(fmt.Errorf("Failed to GET from cluster member with address %q: %w", targetClientURL.String(), err))
		}

		// Write the tokens to the cookies and redirect to the UI
		return response.ManualResponse(func(w http.ResponseWriter) error {
			err := siteManagerState.OIDCVerifier.WriteTokenToCookies(w, tokens.IDToken, tokens.RefreshToken)
			if err != nil {
				return err
			}

			http.Redirect(w, r, stateToken.RedirectURL, http.StatusMovedPermanently)
			return nil
		})
	}

	// If the member address in the state token is the same as the current member address, then handle the callback locally
	callbackHandler := func(w http.ResponseWriter) error {
		siteManagerState.OIDCVerifier.Callback(w, r, stateToken.RedirectURL)
		return nil
	}

	return response.ManualResponse(callbackHandler)
}

func handleForwardedCallback(siteManagerState *state.SiteManagerState, r *http.Request) response.Response {
	var payload types.AuthCallbackPost
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		return response.SmartError(err)
	}

	// For a forwarded request, we need to align the request host with the oidc verifier host so that we can ensure we don't reinitiate the oidc relying party
	r.Host = siteManagerState.OIDCVerifier.Host()

	// Add all the cookies from the source request to the current request so that the oidc relying party can use the same cookies
	// FIXME: we can get cookies from request header instead once microcluster is updated to allow adjusting request headers
	for _, cookie := range payload.Cookies {
		r.AddCookie(&cookie)
	}

	callbackHandler := func(w http.ResponseWriter) error {
		siteManagerState.OIDCVerifier.Callback(w, r, "")
		return nil
	}

	return response.ManualResponse(callbackHandler)
}

func oidcLogout(s *state.SiteManagerState) types.EndpointHandler {
	return func(innerState microState.State, r *http.Request) response.Response {
		redirectURL := r.URL.Query().Get("next")

		logoutHandler := func(w http.ResponseWriter) error {
			s.OIDCVerifier.Logout(w, r, redirectURL)
			return nil
		}

		return response.ManualResponse(logoutHandler)
	}
}

package api

import (
	"context"
	"net/http"

	"github.com/canonical/lxd/lxd/response"
	"github.com/canonical/microcluster/client"
	"github.com/canonical/microcluster/rest/access"
	microState "github.com/canonical/microcluster/state"

	"github.com/canonical/lxd-cluster-manager/internal/api/types"
	"github.com/canonical/lxd-cluster-manager/internal/oidc"
	"github.com/canonical/lxd-cluster-manager/internal/state"
)

func authHandler(clusterManagerState *state.ClusterManagerState) types.AccessHandler {
	return func(clusterState microState.State, r *http.Request) (bool, response.Response) {
		// always allow unix socket requests
		if r.RemoteAddr == "@" {
			return true, nil
		}

		if r.TLS == nil {
			return false, response.Forbidden(nil)
		}

		if clusterManagerState.OIDCVerifier == nil {
			return false, response.Forbidden(nil)
		}

		result, resp, err := clusterManagerState.OIDCVerifier.Auth(r.Context(), r)
		if err != nil {
			// check if the request is a cluster notification with valid cert
			if client.IsNotification(r) && r.TLS != nil {
				hostAddress := clusterState.Address().URL.Host
				trustedCerts := clusterState.Remotes().CertificatesNative()
				trusted, _ := access.Authenticate(clusterState, r, hostAddress, trustedCerts)
				if trusted {
					return true, nil
				}
			}

			if r.URL.Path == "/1.0" {
				return false, response.SyncResponse(false, types.APIRoot{
					Trusted: false,
				})
			}

			return false, resp
		}

		setUserInfoInRequest(result, r)

		return true, resp
	}
}

func setUserInfoInRequest(authResult *oidc.AuthenticationResult, r *http.Request) {
	userInfo := &types.UserInfo{
		Email: authResult.Email,
		Name:  authResult.Name,
	}

	userInfoCtx := context.WithValue(r.Context(), types.UserInfoKey, userInfo)
	*r = *r.WithContext(userInfoCtx)
}

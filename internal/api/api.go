package api

import (
	"context"
	"fmt"
	"net/http"

	"github.com/canonical/lxd/lxd/request"
	"github.com/canonical/lxd/lxd/response"
	"github.com/canonical/lxd/lxd/util"
	"github.com/canonical/microcluster/client"
	"github.com/canonical/microcluster/rest/access"
	microState "github.com/canonical/microcluster/state"

	"github.com/canonical/lxd-cluster-manager/internal/api/types"
	"github.com/canonical/lxd-cluster-manager/internal/oidc"
	"github.com/canonical/lxd-cluster-manager/internal/state"
)

const ctxRemoteClusterID request.CtxKey = "remote-cluster-id"

func oidcAuthHandler(clusterManagerState *state.ClusterManagerState) types.AccessHandler {
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

// mtlsAuthHandler is an access handler that checks the client certificate against trusted remote cluster certificates.
// the client certificate is checked against the trusted certificates in the CertificatesCache which expires every 60 seconds and will be rebuilt with db data.
// if the client certificate is trusted, the certificate fingerprint and the associated remote cluster ID are set in the request context.
// this access handler should be used for protected communication between lxd clusters and lxc cluster manager.
func mtlsAuthHandler(clusterManagerState *state.ClusterManagerState) types.AccessHandler {
	return func(clusterState microState.State, r *http.Request) (bool, response.Response) {
		if r.TLS == nil || len(r.TLS.PeerCertificates) == 0 {
			return false, response.Forbidden(fmt.Errorf("tls is required"))
		}

		if len(r.TLS.PeerCertificates) != 1 {
			return false, response.Forbidden(fmt.Errorf("expected exactly one peer certificate"))
		}

		if clusterManagerState.CertificatesCache.Expired() {
			err := clusterManagerState.CertificatesCache.RebuildCache(r.Context(), clusterState)

			if err != nil {
				return false, response.SmartError(err)
			}
		}

		peerCert := r.TLS.PeerCertificates[0]
		trustedCerts := clusterManagerState.CertificatesCache.GetTrustedCerts()
		trusted, fingerprint := util.CheckMutualTLS(*peerCert, trustedCerts)

		if !trusted {
			return false, response.Forbidden(fmt.Errorf("invalid cluster certificate"))
		}

		remoteClusterCert, _ := clusterManagerState.CertificatesCache.GetCertificateEntry(fingerprint)
		request.SetCtxValue(r, request.CtxUsername, fingerprint)
		request.SetCtxValue(r, ctxRemoteClusterID, remoteClusterCert.ClusterID)

		return true, nil
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

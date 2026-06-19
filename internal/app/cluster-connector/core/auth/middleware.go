package auth

import (
	"crypto/x509"
	"net/http"

	"github.com/canonical/lxd/lxd/response"
	"github.com/canonical/lxd/lxd/util"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/logger"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/types"
	"github.com/gorilla/mux"
)

// AuthMiddleware is a middleware function that checks if the request has a client cert from a remote cluster.
func AuthMiddleware(rc types.RouteConfig) mux.MiddlewareFunc {
	verifier, ok := rc.Auth.Authenticator.(*MtlsAuthenticator)

	middlewareFunc := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if verifier == nil || !ok {
				err := response.Forbidden(nil).Render(w, r)
				if err != nil {
					logger.Log.Errorw("Failed rendering forbidden response due to invalid verifier: %w", err)
				}
				return
			}

			err := verifier.Auth(r.Context(), w, r)
			if err != nil {
				err := response.Forbidden(nil).Render(w, r)
				if err != nil {
					logger.Log.Errorw("Failed rendering forbidden response due to authentication error: %w", err)
				}
				return
			}

			// If auth is successful, then we can proceed
			next.ServeHTTP(w, r)
		})
	}

	return middlewareFunc
}

// InternalAuthMiddleware is a middleware function that checks if the client cert is from the management api.
func InternalAuthMiddleware(rc types.RouteConfig) mux.MiddlewareFunc {
	middlewareFunc := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.TLS == nil || len(r.TLS.PeerCertificates) == 0 {
				logger.Log.Errorw("AUTHN no peer certificate presented for mTLS authentication")
				err := response.Forbidden(nil).Render(w, r)
				if err != nil {
					logger.Log.Errorw("Failed rendering forbidden response due to authentication error", "error", err)
				}
				return
			}

			if len(r.TLS.PeerCertificates) != 1 {
				logger.Log.Errorw("AUTHN more than one peer certificates presented for mTLS authentication")
				err := response.Forbidden(nil).Render(w, r)
				if err != nil {
					logger.Log.Errorw("Failed rendering forbidden response due to authentication error", "error", err)
				}
				return
			}

			managementApiFingerprint := rc.Env.ManagementAPICert.Fingerprint()
			managementApiPublicKey, err := rc.Env.ManagementAPICert.PublicKeyX509()
			if err != nil {
				logger.Log.Errorw("AUTHN failed to parse management API certificate public key", "error", err)
				err = response.Forbidden(nil).Render(w, r)
				if err != nil {
					logger.Log.Errorw("Failed rendering forbidden response due to authentication error", "error", err)
				}
				return
			}

			peerCert := r.TLS.PeerCertificates[0]
			trustedCerts := map[string]x509.Certificate{
				managementApiFingerprint: *managementApiPublicKey,
			}
			trusted, _ := util.CheckMutualTLS(*peerCert, trustedCerts)

			if !trusted {
				logger.Log.Info("AUTHN untrusted peer certificate presented for mTLS")
				err = response.Forbidden(nil).Render(w, r)
				if err != nil {
					logger.Log.Errorw("Failed rendering forbidden response due to authentication error", "error", err)
				}
				return
			}

			logger.Log.Info("AUTHN peer certificate for mTLS authenticated successfully")

			// If auth is successful, then we can proceed
			next.ServeHTTP(w, r)
		})
	}

	return middlewareFunc
}

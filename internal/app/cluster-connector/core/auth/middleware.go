package auth

import (
	"net/http"

	"github.com/canonical/lxd/lxd/response"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/logger"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/types"
	"github.com/gorilla/mux"
)

// AuthMiddleware is a middleware function that checks if the request is authenticated.
func AuthMiddleware(rc types.RouteConfig) mux.MiddlewareFunc {
	verifier, ok := rc.Auth.(*MtlsAuthenticator)

	middlewareFunc := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if verifier == nil || !ok {
				err := response.Forbidden(nil).Render(w, r)
				if err != nil {
					logger.Log.Errorw("Failed rendering forbidden response due to invalid verifier: %w", err)
				}
				return
			}

			_, err := verifier.Auth(r.Context(), w, r)
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

package auth

import (
	"fmt"
	"net/http"

	"github.com/canonical/lxd/lxd/response"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/logger"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/types"
	"github.com/gorilla/mux"
)

// AuthMiddleware is a middleware function that checks if the request is authenticated.
func AuthMiddleware(rc types.RouteConfig) mux.MiddlewareFunc {
	verifier, ok := rc.Auth.Authenticator.(*Verifier)

	middlewareFunc := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.TLS == nil {
				err := response.Forbidden(fmt.Errorf("TLS is required")).Render(w, r)
				if err != nil {
					logger.Log.Errorf("Failed rendering forbidden response due to missing TLS: %v", err)
				}
				return
			}

			if verifier == nil || !ok {
				err := response.InternalError(nil).Render(w, r)
				if err != nil {
					logger.Log.Errorf("Failed rendering internal server error response due to invalid verifier: %v", err)
				}
				return
			}

			err := verifier.Auth(r.Context(), w, r)
			if err != nil {
				err := response.Unauthorized(nil).Render(w, r)
				if err != nil {
					logger.Log.Errorf("Failed rendering unauthorized response due to authentication error: %v", err)
				}
				return
			}

			// If auth is successful, then we can proceed
			next.ServeHTTP(w, r)
		})
	}

	return middlewareFunc
}

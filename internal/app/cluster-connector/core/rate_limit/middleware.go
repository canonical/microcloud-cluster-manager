package rate_limit

import (
	"net/http"

	"github.com/canonical/lxd/lxd/response"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/logger"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/types"
	"github.com/gorilla/mux"
)

func RateLimitMiddleware(rc types.RouteConfig) mux.MiddlewareFunc {
	rateLimiter, ok := rc.RateLimiter.(*RateLimiter)
	middlewareFunc := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if rateLimiter == nil || !ok {
				err := response.InternalError(nil).Render(w, r)
				if err != nil {
					logger.Log.Errorw("Failed rendering Internal Server Error response due to invalid rateLimiter: %w", err)
				}
				return
			}

			allowRequest, err := rateLimiter.CheckLimit(r.Context(), w, r)
			if err != nil {
				err := response.InternalError(nil).Render(w, r)
				if err != nil {
					logger.Log.Errorw("Failed rendering Internal Server Error response due to rateLimiter error: %w", err)
				}
				return
			}

			if !allowRequest {
				http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
	return middlewareFunc
}

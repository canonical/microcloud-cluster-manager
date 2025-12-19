package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/canonical/lxd/lxd/response"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/logger"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/request"
	"github.com/google/uuid"
)

// RequestTrace is a middleware that adds a trace ID and timestamp to the request context.
func RequestTrace(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, err := uuid.NewUUID()
		if err != nil {
			err := response.InternalError(err).Render(w, r)
			if err != nil {
				logger.Log.Errorw("Failed rendering internal error response due to failed UUID generation: %w", err)
			}
			return
		}

		v := request.Values{
			TraceID: id.String(),
			Now:     time.Now(),
		}
		ctx := r.Context()
		ctx = context.WithValue(ctx, request.RequestKey(), &v)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// LogRequest is a middleware that logs a request/response cycle.
func LogRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		// If the context is missing this value, we can't log anything
		v, err := request.GetValues(ctx)
		if err != nil {
			err := response.InternalError(err).Render(w, r)
			if err != nil {
				logger.Log.Errorw("Failed rendering internal error response due to missing request values: %w", err)
			}
			return
		}

		// Generate a new trace ID
		logger.Log.Infow(
			"request started",
			"traceid", v.TraceID,
			"method", r.Method,
			"path", r.URL.Path,
			"remoteaddr", r.RemoteAddr,
		)

		// Create a custom response writer to capture status code
		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(rw, r)

		logger.Log.Infow(
			"request completed",
			"traceid", v.TraceID,
			"method", r.Method,
			"path", r.URL.Path,
			"remoteaddr", r.RemoteAddr,
			"statuscode", rw.statusCode,
			"since", time.Since(v.Now),
		)
	})
}

// Custom response writer to capture status code.
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader captures the status code.
func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

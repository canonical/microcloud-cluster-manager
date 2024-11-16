package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/canonical/lxd-cluster-manager/internal/pkg/request"
	"github.com/canonical/lxd/lxd/response"
	"github.com/google/uuid"
)

// RequestTrace is a middleware that adds a trace ID and timestamp to the request context
func (m *Middleware) RequestTrace(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, err := uuid.NewUUID()
		if err != nil {
			response.InternalError(err).Render(w, r)
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

// LogRequest is a middleware that logs a request/response cycle
func (m *Middleware) LogRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		// If the context is missing this value, we can't log anything
		v, err := request.GetValues(ctx)
		if err != nil {
			response.InternalError(err).Render(w, r)
			return
		}

		// Generate a new trace ID
		m.log.Infow(
			"request started",
			"traceid", v.TraceID,
			"method", r.Method,
			"path", r.URL.Path,
			"remoteaddr", r.RemoteAddr,
		)

		// Create a custom response writer to capture status code
		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(rw, r)

		m.log.Infow(
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

// Custom response writer to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader captures the status code
func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

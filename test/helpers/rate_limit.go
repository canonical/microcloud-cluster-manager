package helpers

import (
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/canonical/microcloud-cluster-manager/internal/app/cluster-connector/core/rate_limit"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/types"
	"golang.org/x/time/rate"
)

// getRateLimiter creates and returns a RateLimiter for testing.
func getRateLimiter(refillRate rate.Limit, bucketSize int, maxClients int, clientActiveInterval time.Duration, cleanupInterval time.Duration, logInterval time.Duration) *rate_limit.RateLimiter {
	return rate_limit.NewRateLimiter(refillRate, bucketSize, clientActiveInterval, maxClients, cleanupInterval, logInterval)
}

// Mock handler that just returns 200 OK.
func mockHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})
}

// GetHandlerWithRateLimiting returns an HTTP handler wrapped with Token Bucket rate limiting middleware.
func GetHandlerWithRateLimiting(refillRate rate.Limit, bucketSize int, maxClients int, ttl time.Duration, cleanupInterval time.Duration, logInterval time.Duration) http.Handler {
	rl := getRateLimiter(refillRate, bucketSize, maxClients, ttl, cleanupInterval, logInterval)
	middleware := rate_limit.RateLimitMiddleware(types.RouteConfig{
		RateLimiter: rl,
	})
	return middleware(mockHandler())
}

func GetRandomIP() string {
	return fmt.Sprintf("%d.%d.%d.%d:%d", rand.Intn(256), rand.Intn(256), rand.Intn(256), rand.Intn(256), rand.Intn(65536))
}

// SendTestRequest sends a test HTTP request to the provided handler with the specified client IP.
func SendTestRequest(handler http.Handler, clientIP string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = clientIP
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	return rr
}

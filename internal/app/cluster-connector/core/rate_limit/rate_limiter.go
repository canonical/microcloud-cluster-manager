package rate_limit

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/canonical/microcloud-cluster-manager/internal/pkg/logger"
	"golang.org/x/time/rate"
)

type ClientLimiter struct {
	limiter    *rate.Limiter
	lastSeen   time.Time
	lastLogged time.Time
}

// RateLimiter manages per-client limiters.
type RateLimiter struct {
	mu                   sync.Mutex
	refillRate           rate.Limit    // tokens added to bucket per second
	bucketSize           int           // maximum tokens in the bucket (requests/second)
	clientActiveInterval time.Duration // duration after which inactive clients are considered for cleanup
	maxClients           int
	clients              map[string]*ClientLimiter
	cleanupInterval      time.Duration // duration between cleanup runs
	logInterval          time.Duration // duration between logs to avoid I/O strain
	lastLogged           time.Time     // global last logged time
}

// NewRateLimiter creates a new RateLimiter.
func NewRateLimiter(refillRate rate.Limit, bucketSize int, clientActiveInterval time.Duration, maxClients int, cleanupInterval time.Duration, logInterval time.Duration) *RateLimiter {
	rl := &RateLimiter{
		clients:              make(map[string]*ClientLimiter),
		refillRate:           refillRate,
		bucketSize:           bucketSize,
		clientActiveInterval: clientActiveInterval,
		maxClients:           maxClients,
		cleanupInterval:      cleanupInterval,
		logInterval:          logInterval,
		lastLogged:           time.Now(),
	}

	// Start cleanup goroutine
	go rl.cleanupLoop()
	return rl
}

// CheckLimit checks that the request from the client is within the rate limit.
func (rl *RateLimiter) CheckLimit(ctx context.Context, w http.ResponseWriter, r *http.Request) (bool, error) {
	clientID := extractClientIP(r)
	lim := rl.getOrCreateClientLimiter(clientID)

	if lim == nil || lim.limiter == nil {
		logger.Log.Info("Could not create rate limiter for client %s", clientID)
		return false, fmt.Errorf("could not create rate limiter for client %s", clientID)
	}

	if !lim.limiter.Allow() {
		if time.Since(lim.lastLogged) >= rl.logInterval {
			lim.lastLogged = time.Now()
			logger.Log.Info("Rate limit exceeded for client %s", clientID)
		}
		w.Header().Set("Retry-After", getRetryAfterHeader(lim.limiter))
		return false, nil
	}

	return true, nil
}

// getOrCreateClientLimiter returns an existing limiter for client or creates a new one.
func (rl *RateLimiter) getOrCreateClientLimiter(key string) *ClientLimiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	if cl, exists := rl.clients[key]; exists {
		cl.lastSeen = time.Now()
		return cl
	}

	if len(rl.clients) >= rl.maxClients {
		if time.Since(rl.lastLogged) >= rl.logInterval {
			rl.lastLogged = time.Now()
			logger.Log.Info("Rate limiter store is full: maximum %d clients reached", rl.maxClients)
		}
		return nil
	}

	lim := rate.NewLimiter(rl.refillRate, rl.bucketSize)
	rl.clients[key] = &ClientLimiter{
		limiter:    lim,
		lastSeen:   time.Now(),
		lastLogged: time.Now(),
	}

	return rl.clients[key]
}

// cleanupLoop periodically deletes old entries to prevent memory leak.
func (rl *RateLimiter) cleanupLoop() {
	ticker := time.NewTicker(rl.cleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		rl.cleanup()
	}
}

// cleanup removes clients not seen for clientActiveInterval duration.
func (rl *RateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	for key, cl := range rl.clients {
		if time.Since(cl.lastSeen) > rl.clientActiveInterval {
			delete(rl.clients, key)
		}
	}
}

// extractClientIP tries to get the real client IP.
func extractClientIP(r *http.Request) string {
	// X-Forwarded-For when behind proxies
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		return xff
	}

	// Fallback to RemoteAddr
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}

	return ip
}

// getRetryAfterHeader calculates the Retry-After header value based on the limiter's rate.
func getRetryAfterHeader(limiter *rate.Limiter) string {
	// Calculate time until next token based on rate
	// If rate is N tokens per second, each token takes 1/N seconds
	limit := limiter.Limit()
	if limit <= 0 {
		return "60" // Default fallback
	}

	secondsPerToken := 1.0 / float64(limit)
	retryAfter := int(secondsPerToken) + 1

	return fmt.Sprintf("%d", retryAfter)
}

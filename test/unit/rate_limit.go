package main

import (
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/canonical/microcloud-cluster-manager/test/helpers"
)

func testRateLimitMiddleware_AllowsRequests() (testName string, testFunc func(t *testing.T)) {
	return "rate limit middleware allows requests", func(t *testing.T) {
		var condition string

		{
			condition = "Should allow request with generous rate limit"

			handler := helpers.GetHandlerWithRateLimiting(10, 20, 1000, 1*time.Minute, 1*time.Minute, 1*time.Minute)
			clientIP := helpers.GetRandomIP()
			rr := helpers.SendTestRequest(handler, clientIP)

			var err error
			if rr.Code != http.StatusOK {
				err = fmt.Errorf("expected status OK, got %d", rr.Code)
			} else if rr.Body.String() != "OK" {
				err = fmt.Errorf("expected body 'OK', got '%s'", rr.Body.String())
			}

			helpers.LogTestOutcome(t, condition, err)
		}
	}
}

func testRateLimitMiddleware_BlocksExcessiveRequests() (testName string, testFunc func(t *testing.T)) {
	return "rate limit middleware blocks excessive requests", func(t *testing.T) {
		var condition string

		{
			condition = "Should block second request with rate limiting when tokens are exhausted"

			handler := helpers.GetHandlerWithRateLimiting(1, 1, 1000, 1*time.Minute, 1*time.Minute, 1*time.Minute)
			clientIP := helpers.GetRandomIP()
			helpers.SendTestRequest(handler, clientIP)        // First request
			rr2 := helpers.SendTestRequest(handler, clientIP) // Second request

			var err error
			if rr2.Code != http.StatusTooManyRequests {
				err = fmt.Errorf("second request: expected status 429, got %d", rr2.Code)
			}

			helpers.LogTestOutcome(t, condition, err)
		}
	}
}

func testRateLimitMiddleware_MaxClients() (testName string, testFunc func(t *testing.T)) {
	return "rate limit middleware enforces max clients", func(t *testing.T) {
		var condition string

		{
			condition = "Should handle many different clients up to limit"

			handler := helpers.GetHandlerWithRateLimiting(10, 20, 50, 10*time.Minute, 10*time.Minute, 1*time.Minute)
			numClients := 100
			successCount := 50

			for i := 0; i < numClients; i++ {
				clientIP := helpers.GetRandomIP()
				rr := helpers.SendTestRequest(handler, clientIP)

				if rr.Code == http.StatusOK {
					successCount--
				}
			}

			var err error
			if successCount < 0 {
				err = errors.New("too many successful requests, exceeded expected limit")
			} else if successCount > 0 {
				err = fmt.Errorf("not enough successful requests, expected at least %d successes", 50)
			}

			helpers.LogTestOutcome(t, condition, err)
		}
	}
}

func testRateLimitMiddleware_BucketSizeAllowance() (testName string, testFunc func(t *testing.T)) {
	return "rate limit middleware allows requests within bucket size", func(t *testing.T) {
		var condition string

		{
			condition = "Should allow 3 requests then block"

			handler := helpers.GetHandlerWithRateLimiting(1, 3, 1000, 1*time.Minute, 1*time.Minute, 1*time.Minute)
			clientIP := helpers.GetRandomIP()

			for i := 0; i < 3; i++ {
				helpers.SendTestRequest(handler, clientIP)
			}

			rr := helpers.SendTestRequest(handler, clientIP)

			var err error
			if rr.Code != http.StatusTooManyRequests {
				err = fmt.Errorf("4th request: expected status 429, got %d", rr.Code)
			}

			helpers.LogTestOutcome(t, condition, err)
		}
	}
}

func testRateLimitMiddleware_TokenRefill() (testName string, testFunc func(t *testing.T)) {
	return "rate limit middleware refills tokens over time", func(t *testing.T) {
		var condition string

		{
			condition = "Should allow request after token refill"

			bucketSize := 5
			handler := helpers.GetHandlerWithRateLimiting(1, bucketSize, 50, 2*time.Minute, 2*time.Minute, 1*time.Minute)
			clientIP := helpers.GetRandomIP()

			for i := 0; i < bucketSize; i++ {
				helpers.SendTestRequest(handler, clientIP)
			}

			time.Sleep(1 * time.Second) // Wait for some time to allow token refill
			rr2 := helpers.SendTestRequest(handler, clientIP)

			var err error
			if rr2.Code != http.StatusOK {
				err = fmt.Errorf("after refill: expected status OK, got %d", rr2.Code)
			}

			helpers.LogTestOutcome(t, condition, err)
		}
	}
}

func testRateLimitMiddleware_CleanupLoop() (testName string, testFunc func(t *testing.T)) {
	return "rate limit middleware cleanup loop removes old clients", func(t *testing.T) {
		var condition string

		{
			condition = "Should remove inactive clients after cleanup interval"

			// Use short cleanup interval for testing
			cleanupInterval := 2 * time.Second
			clientTTL := 1 * time.Second
			handler := helpers.GetHandlerWithRateLimiting(10, 20, 1000, clientTTL, clientTTL, cleanupInterval)

			// Create a client and make a request
			clientIP := helpers.GetRandomIP()
			rr1 := helpers.SendTestRequest(handler, clientIP)

			var err error
			if rr1.Code != http.StatusOK {
				err = fmt.Errorf("initial request: expected status OK, got %d", rr1.Code)
			}

			// Wait for client TTL to expire and cleanup to run
			time.Sleep(clientTTL + cleanupInterval + 500*time.Millisecond)

			// Make another request - should succeed with fresh bucket if cleanup worked
			rr2 := helpers.SendTestRequest(handler, clientIP)
			if rr2.Code != http.StatusOK {
				err = fmt.Errorf("after cleanup: expected status OK, got %d", rr2.Code)
			}

			helpers.LogTestOutcome(t, condition, err)
		}
	}
}

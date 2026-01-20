package main

import (
	"testing"

	"github.com/canonical/microcloud-cluster-manager/test/types"
)

var tests = []types.UnitTest{
	testRateLimitMiddleware_AllowsRequests,
	testRateLimitMiddleware_BlocksExcessiveRequests,
	testRateLimitMiddleware_MaxClients,
	testRateLimitMiddleware_BucketSizeAllowance,
	testRateLimitMiddleware_TokenRefill,
	testRateLimitMiddleware_CleanupLoop,
}

func TestUnit(t *testing.T) {
	// run tests
	for _, tt := range tests {
		testName, testFunc := tt()
		t.Run(testName, testFunc)
	}
}

package main

import (
	"testing"

	"github.com/canonical/microcloud-cluster-manager/test/helpers"
	"github.com/canonical/microcloud-cluster-manager/test/types"
)

var tests = []types.Test{
	testRemoteClusterSuccess,
	testRemoteClusterSuccessWithMetrics,
	testRemoteClusterJoinInvalid,
	testRemoteClusterJoinExpiredToken,
	testRemoteClusterStatusNoCert,
	testRemoteClusterStatusInvalidCert,
	testAuthAdminUserAllowsAccess,
	testAuthNonAdminDenyAccess,
	testAuthNonAdminUnprotectedEndpointAllowAccess,
	testAuthNoCookiesDeniesAccess,
	testAuthTamperedIDTokenDeniesAccess,
	testAuthTamperedRefreshTokenDeniesAccess,
	testAuthUnknownSessionIDDeniesAccess,
	testAuthOnlySessionIDCookieDeniesAccess,
	//testAuthLogoutInvalidatesCookies, //TODO: re-enable once we reliably log out users.
	testAuthOIDCLoginRedirects,
	testAuthOIDCCallbackWithInvalidStateReturns500,
}

func TestE2E(t *testing.T) {
	env := initE2E(t)

	// cleaning up environment
	defer func() {
		err := env.Cleanup()
		if err != nil {
			t.Fatalf("Failed to cleanup environment: %v", err)
		}
	}()

	// run tests
	for _, tt := range tests {
		testName, testFunc := tt(env)
		t.Run(testName, testFunc)
	}
}

func initE2E(t *testing.T) *helpers.Environment {
	env := helpers.NewEnv()
	err := env.Init()
	helpers.LogTestOutcome(t, "Initialize environment", err)
	return env
}

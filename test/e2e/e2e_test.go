package main

import (
	"testing"

	"github.com/canonical/lxd-cluster-manager/test/helpers"
	"github.com/canonical/lxd-cluster-manager/test/types"
)

var tests = []types.Test{
	testRemoteClusterSuccess,
	testRemoteClusterJoinInvalid,
	testRemoteClusterJoinExpiredToken,
	testRemoteClusterStatusNoCert,
	testRemoteClusterStatusInvalidCert,
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

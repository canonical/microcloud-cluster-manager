package main

import (
	"testing"
	"time"

	"github.com/canonical/lxd-cluster-manager/test/helpers"
)

func testRemoteClusterJoinInvalidHMAC(env *helpers.Environment) (testName string, testFunc func(t *testing.T)) {
	return "lxd remote cluster join with invalid HMAC", func(t *testing.T) {
		remoteClusterName := "remote_cluster_join_invalid_hmac"
		var condition string

		{
			condition = "Should fail join request validation with invalid HMAC"

			tokenData, err := helpers.CreateAndReturnRemoteClusterJoinToken(env, remoteClusterName, time.Time{})
			if err != nil {
				helpers.LogTestOutcome(t, condition, err)
			}

			tokenData.Secret = "invalid_secret"
			err = sendJoinRequest(env, tokenData)
			if err != nil && err.Error() == "not authorized" {
				err = nil
			}

			helpers.LogTestOutcome(t, condition, err)
		}
	}
}

func testRemoteClusterJoinExpiredToken(env *helpers.Environment) (testName string, testFunc func(t *testing.T)) {
	return "lxd remote cluster join with expired token", func(t *testing.T) {
		remoteClusterName := "remote_cluster_join_expired_token"
		var condition string

		{
			condition = "Should fail join request validation with expired token"

			expiry := time.Now().Add(1 * time.Second)
			tokenData, err := helpers.CreateAndReturnRemoteClusterJoinToken(env, remoteClusterName, expiry)
			if err != nil {
				helpers.LogTestOutcome(t, condition, err)
			}

			// Ensure token expires before sending join request
			time.Sleep(1 * time.Second)
			err = sendJoinRequest(env, tokenData)
			if err != nil && err.Error() == "token has expired" {
				err = nil
			}

			helpers.LogTestOutcome(t, condition, err)
		}
	}
}

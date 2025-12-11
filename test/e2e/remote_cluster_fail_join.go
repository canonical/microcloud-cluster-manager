package main

import (
	"errors"
	"testing"
	"time"

	"github.com/canonical/microcloud-cluster-manager/test/helpers"
)

func testRemoteClusterJoinInvalid(env *helpers.Environment) (testName string, testFunc func(t *testing.T)) {
	return "lxd remote cluster join with invalid secret", func(t *testing.T) {
		remoteClusterName := helpers.GetRandomName("remote_cluster_join_invalid_secret")
		var condition string

		{
			condition = "Should fail join request validation with invalid secret"

			tokenData, err := helpers.CreateAndReturnRemoteClusterJoinToken(env, remoteClusterName, time.Time{})
			if err != nil {
				helpers.LogTestOutcome(t, condition, err)
			}

			tokenData.Secret = "invalid_secret"
			err = helpers.SendJoinRequest(env, tokenData)
			if err != nil && err.Error() == "Forbidden" {
				err = nil
			} else {
				err = errors.New("expected forbidden error not received")
			}

			helpers.LogTestOutcome(t, condition, err)
		}

		env.RemoveRemoteClusterToken(remoteClusterName)
		env.RemoveRemoteCluster(remoteClusterName)
	}
}

func testRemoteClusterJoinExpiredToken(env *helpers.Environment) (testName string, testFunc func(t *testing.T)) {
	return "lxd remote cluster join with expired token", func(t *testing.T) {
		remoteClusterName := helpers.GetRandomName("remote_cluster_join_expired_token")
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
			err = helpers.SendJoinRequest(env, tokenData)
			if err != nil && err.Error() == "tokenFromDb has expired" {
				err = nil
			} else {
				err = errors.New("expected token expired error not received")
			}

			helpers.LogTestOutcome(t, condition, err)
		}

		env.RemoveRemoteClusterToken(remoteClusterName)
		env.RemoveRemoteCluster(remoteClusterName)
	}
}

package main

import (
	"testing"
	"time"

	"github.com/canonical/lxd-site-manager/test/helpers"
)

func testSiteJoinInvalidHMAC(env *helpers.Environment) (testName string, testFunc func(t *testing.T)) {
	return "lxd site join with invalid HMAC", func(t *testing.T) {
		siteName := "site_join_invalid_hmac"
		var condition string

		{
			condition = "Should fail join request validation with invalid HMAC"

			tokenData, err := helpers.CreateAndReturnSiteJoinToken(env, siteName, time.Time{})
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

func testSiteJoinExpiredToken(env *helpers.Environment) (testName string, testFunc func(t *testing.T)) {
	return "lxd site join with expired token", func(t *testing.T) {
		siteName := "site_join_expired_token"
		var condition string

		{
			condition = "Should fail join request validation with expired token"

			expiry := time.Now().Add(1 * time.Second)
			tokenData, err := helpers.CreateAndReturnSiteJoinToken(env, siteName, expiry)
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

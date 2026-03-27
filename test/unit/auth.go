package main

import (
	"fmt"
	"testing"

	"github.com/canonical/microcloud-cluster-manager/test/helpers"
)

func testAuthorizor_CheckPermissions_AdminUser() (testName string, testFunc func(t *testing.T)) {
	return "Authorizor CheckPermissions allows admin user", func(t *testing.T) {
		var condition string

		{
			condition = "Should allow admin user to pass permission check"

			authorizor, err := helpers.GetManagementAPIAuthorizor()
			if err != nil {
				helpers.LogTestOutcome(t, condition, fmt.Errorf("failed to create authorizor: %w", err))
				return
			}

			ctx := helpers.GetContextWithUserInfo(true) // Admin user
			err = authorizor.CheckPermissions(ctx, []string{})
			if err != nil {
				err = fmt.Errorf("admin user should have full access, but got error: %w", err)
			}
			helpers.LogTestOutcome(t, condition, err)
		}
	}
}

func testAuthorizor_CheckPermissions_NonAdminUser() (testName string, testFunc func(t *testing.T)) {
	return "Authorizor CheckPermissions allows non-admin user", func(t *testing.T) {
		var condition string

		{
			condition = "Should not allow non-admin user to pass permission check"

			authorizor, err := helpers.GetManagementAPIAuthorizor()
			if err != nil {
				helpers.LogTestOutcome(t, condition, fmt.Errorf("failed to create authorizor: %w", err))
				return
			}

			ctx := helpers.GetContextWithUserInfo(false) // Non-admin user
			err = authorizor.CheckPermissions(ctx, []string{})
			if err != nil {
				err = nil
			} else {
				err = fmt.Errorf("non-admin user should not have access, but permission check passed")
			}
			helpers.LogTestOutcome(t, condition, err)
		}
	}
}

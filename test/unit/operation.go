package main

import (
	"errors"
	"fmt"
	"testing"

	"github.com/canonical/microcloud-cluster-manager/internal/app/cluster-connector/core/operation"
	"github.com/canonical/microcloud-cluster-manager/test/helpers"
)

func testOperation_Paths() (testName string, testFunc func(t *testing.T)) {
	return "Operation paths escape the operation ID", func(t *testing.T) {
		condition := "LXD operation paths should escape user input"

		cases := []struct{ got, want string }{
			{operation.Path("2f1b4e7a-6c3d-4f9e-8a21-5d0c7b3e9f10"), "/1.0/operations/2f1b4e7a-6c3d-4f9e-8a21-5d0c7b3e9f10"},
			{operation.Path("../instances?x=1"), "/1.0/operations/..%2Finstances%3Fx=1"},
			{operation.Path("op#frag"), "/1.0/operations/op%23frag"},
		}

		var err error
		for _, c := range cases {
			if c.got != c.want {
				err = errors.Join(err, fmt.Errorf("expected path %q, got %q", c.want, c.got))
			}
		}

		helpers.LogTestOutcome(t, condition, err)
	}
}

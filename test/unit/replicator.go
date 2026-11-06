package main

import (
	"errors"
	"fmt"
	"slices"
	"testing"

	"github.com/canonical/microcloud-cluster-manager/internal/app/cluster-connector/core/replicator"
	"github.com/canonical/microcloud-cluster-manager/test/helpers"
)

func testReplicator_ValidateAction() (testName string, testFunc func(t *testing.T)) {
	return "Replicator ValidateAction accepts only supported run actions", func(t *testing.T) {
		condition := "start and restore should be accepted, any other action rejected"

		var err error
		for _, action := range []string{"start", "restore"} {
			if replicator.ValidateAction(action) != nil {
				err = fmt.Errorf("action %q should be accepted", action)
				break
			}
		}

		for _, action := range []string{"", "stop", "Start", "start "} {
			if err != nil {
				break
			}

			if replicator.ValidateAction(action) == nil {
				err = fmt.Errorf("action %q should be rejected", action)
			}
		}

		helpers.LogTestOutcome(t, condition, err)
	}
}

func testReplicator_Requirements() (testName string, testFunc func(t *testing.T)) {
	return "Replicator Requirements require the replicators API extension", func(t *testing.T) {
		condition := "Replicator requests should require the replicators API extension and leave permissions to LXD"

		var err error
		requirements := replicator.Requirements()
		if !slices.Equal(requirements.APIExtensions, []string{"replicators"}) {
			err = fmt.Errorf("unexpected API extensions: %v", requirements.APIExtensions)
		} else if len(requirements.Permissions) != 0 {
			err = fmt.Errorf("unexpected permissions: %v", requirements.Permissions)
		}

		helpers.LogTestOutcome(t, condition, err)
	}
}

func testReplicator_Paths() (testName string, testFunc func(t *testing.T)) {
	return "Replicator paths escape the project and replicator name", func(t *testing.T) {
		condition := "LXD replicator paths should be scoped to the project and escape user input"

		cases := []struct{ got, want string }{
			{replicator.CollectionPath("default"), "/1.0/replicators?project=default"},
			{replicator.ListPath("default"), "/1.0/replicators?project=default&recursion=2"},
			{replicator.ListPath(""), "/1.0/replicators?all-projects=true&recursion=2"},
			{replicator.Path("default", "backup"), "/1.0/replicators/backup?project=default"},
			{replicator.StatePath("default", "backup"), "/1.0/replicators/backup/state?project=default"},
			{replicator.ListPath("a&recursion=0"), "/1.0/replicators?project=a%26recursion%3D0&recursion=2"},
			{replicator.Path("p", "../instances?x=1"), "/1.0/replicators/..%2Finstances%3Fx=1?project=p"},
			{replicator.StatePath("p#frag", "name/../state"), "/1.0/replicators/name%2F..%2Fstate/state?project=p%23frag"},
		}

		var err error
		for _, c := range cases {
			if got, want := c.got, c.want; got != want {
				err = errors.Join(err, fmt.Errorf("expected path %q, got %q", want, got))
			}
		}

		helpers.LogTestOutcome(t, condition, err)
	}
}

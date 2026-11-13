package main

import (
	"errors"
	"fmt"
	"testing"

	"github.com/canonical/microcloud-cluster-manager/internal/app/cluster-connector/core/project"
	"github.com/canonical/microcloud-cluster-manager/test/helpers"
)

func testProject_Paths() (testName string, testFunc func(t *testing.T)) {
	return "Project paths escape the project name", func(t *testing.T) {
		condition := "LXD project paths should escape user input"

		cases := []struct{ got, want string }{
			{project.ListPath(), "/1.0/projects?recursion=1"},
			{project.CreatePath(), "/1.0/projects"},
			{project.Path("default"), "/1.0/projects/default"},
			{project.Path("../instances?x=1"), "/1.0/projects/..%2Finstances%3Fx=1"},
			{project.Path("p#frag"), "/1.0/projects/p%23frag"},
			{project.StatePath("default", false), "/1.0/projects/default/state"},
			{project.StatePath("default", true), "/1.0/projects/default/state?force=true"},
			{project.StatePath("../instances?x=1", false), "/1.0/projects/..%2Finstances%3Fx=1/state"},
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

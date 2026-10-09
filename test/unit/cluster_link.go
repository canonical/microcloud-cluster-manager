package main

import (
	"fmt"
	"testing"

	clusterLink "github.com/canonical/microcloud-cluster-manager/internal/app/cluster-connector/core/cluster_link"
	"github.com/canonical/microcloud-cluster-manager/test/helpers"
)

func testClusterLink_RequiredPermissions_UnsupportedType() (testName string, testFunc func(t *testing.T)) {
	return "ClusterLink RequiredPermissions rejects unsupported link types", func(t *testing.T) {
		condition := "Unknown or missing link types should not be supported"

		var err error
		for _, linkType := range []string{"", "unidirectional", "public", "bogus"} {
			_, _, ok := clusterLink.RequiredPermissions(linkType)
			if ok {
				err = fmt.Errorf("link type %q should not be supported", linkType)
				break
			}
		}

		helpers.LogTestOutcome(t, condition, err)
	}
}

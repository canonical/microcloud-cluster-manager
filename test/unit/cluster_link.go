package main

import (
	"fmt"
	"slices"
	"testing"

	"github.com/canonical/lxd/shared/api"
	"github.com/canonical/lxd/shared/entity"
	clusterLink "github.com/canonical/microcloud-cluster-manager/internal/app/cluster-connector/core/cluster_link"
	"github.com/canonical/microcloud-cluster-manager/test/helpers"
)

func serverPermission(entitlement string) api.Permission {
	return api.Permission{
		EntityType:      string(entity.TypeServer),
		EntityReference: entity.ServerURL().String(),
		Entitlement:     entitlement,
	}
}

func testClusterLink_RequiredEntitlements_Bidirectional() (testName string, testFunc func(t *testing.T)) {
	return "ClusterLink RequiredEntitlements for bidirectional links", func(t *testing.T) {
		condition := "Bidirectional links should require create-cluster-link and create-identity on both clusters"

		source, target, ok := clusterLink.RequiredEntitlements("bidirectional")

		var err error
		if !ok {
			err = fmt.Errorf("bidirectional link type should be supported")
		} else if !slices.Contains(source, clusterLink.EntitlementCreateClusterLink) || !slices.Contains(source, clusterLink.EntitlementCreateIdentity) {
			err = fmt.Errorf("source entitlements should contain %q and %q, got %v",
				clusterLink.EntitlementCreateClusterLink, clusterLink.EntitlementCreateIdentity, source)
		} else if !slices.Contains(target, clusterLink.EntitlementCreateClusterLink) || !slices.Contains(target, clusterLink.EntitlementCreateIdentity) {
			err = fmt.Errorf("target entitlements should contain %q and %q, got %v",
				clusterLink.EntitlementCreateClusterLink, clusterLink.EntitlementCreateIdentity, target)
		}

		helpers.LogTestOutcome(t, condition, err)
	}
}

func testClusterLink_RequiredEntitlements_UnsupportedType() (testName string, testFunc func(t *testing.T)) {
	return "ClusterLink RequiredEntitlements rejects unsupported link types", func(t *testing.T) {
		condition := "Unknown or missing link types should not be supported"

		var err error
		for _, linkType := range []string{"", "unidirectional", "public", "bogus"} {
			_, _, ok := clusterLink.RequiredEntitlements(linkType)
			if ok {
				err = fmt.Errorf("link type %q should not be supported", linkType)
				break
			}
		}

		helpers.LogTestOutcome(t, condition, err)
	}
}

func testClusterLink_HasServerEntitlement() (testName string, testFunc func(t *testing.T)) {
	return "ClusterLink HasServerEntitlement matches server entity permissions", func(t *testing.T) {
		cases := []struct {
			condition           string
			effectivePermission []api.Permission
			entitlement         clusterLink.Entitlement
			want                bool
		}{
			{
				condition:           "Exact entitlement on the server entity should match",
				effectivePermission: []api.Permission{serverPermission("can_create_cluster_links")},
				entitlement:         clusterLink.EntitlementCreateClusterLink,
				want:                true,
			},
			{
				condition:           "Admin entitlement on the server entity should imply any entitlement",
				effectivePermission: []api.Permission{serverPermission("admin")},
				entitlement:         clusterLink.EntitlementCreateClusterLink,
				want:                true,
			},
			{
				condition:           "A different entitlement on the server entity should not match",
				effectivePermission: []api.Permission{serverPermission("can_view")},
				entitlement:         clusterLink.EntitlementCreateClusterLink,
				want:                false,
			},
			{
				condition: "The same entitlement on another entity should not match",
				effectivePermission: []api.Permission{{
					EntityType:      string(entity.TypeClusterLink),
					EntityReference: entity.ClusterLinkURL("foo").String(),
					Entitlement:     "can_create_cluster_links",
				}},
				entitlement: clusterLink.EntitlementCreateClusterLink,
				want:        false,
			},
			{
				condition:           "Empty permissions should not match",
				effectivePermission: nil,
				entitlement:         clusterLink.EntitlementCreateClusterLink,
				want:                false,
			},
		}

		for _, c := range cases {
			var err error
			if got := clusterLink.HasServerEntitlement(c.effectivePermission, c.entitlement); got != c.want {
				err = fmt.Errorf("expected %v, got %v", c.want, got)
			}

			helpers.LogTestOutcome(t, c.condition, err)
		}
	}
}

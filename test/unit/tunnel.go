package main

import (
	"fmt"
	"testing"

	"github.com/canonical/lxd/shared/api"
	"github.com/canonical/lxd/shared/entity"
	"github.com/canonical/microcloud-cluster-manager/internal/app/cluster-connector/core/tunnel"
	"github.com/canonical/microcloud-cluster-manager/test/helpers"
)

func serverPermission(entitlement string) api.Permission {
	return api.Permission{
		EntityType:      string(entity.TypeServer),
		EntityReference: entity.ServerURL().String(),
		Entitlement:     entitlement,
	}
}

func testTunnel_HasPermission() (testName string, testFunc func(t *testing.T)) {
	return "Tunnel HasPermission matches effective LXD permissions", func(t *testing.T) {
		cases := []struct {
			condition           string
			effectivePermission []api.Permission
			requiredPermission  api.Permission
			want                bool
		}{
			{
				condition:           "Exact entitlement on the server entity should match",
				effectivePermission: []api.Permission{serverPermission("can_create_cluster_links")},
				requiredPermission:  serverPermission("can_create_cluster_links"),
				want:                true,
			},
			{
				condition:           "Admin entitlement on the server entity should imply any entitlement",
				effectivePermission: []api.Permission{serverPermission("admin")},
				requiredPermission:  serverPermission("can_create_cluster_links"),
				want:                true,
			},
			{
				condition: "Admin entitlement on the server entity should imply permissions on other entities",
				effectivePermission: []api.Permission{
					serverPermission("admin"),
				},
				requiredPermission: api.Permission{
					EntityType:      string(entity.TypeInstance),
					EntityReference: "/1.0/instances/c1?project=default",
					Entitlement:     "can_view",
				},
				want: true,
			},
			{
				condition:           "A different entitlement on the server entity should not match",
				effectivePermission: []api.Permission{serverPermission("can_view")},
				requiredPermission:  serverPermission("can_create_cluster_links"),
				want:                false,
			},
			{
				condition: "The same entitlement on another entity should not match",
				effectivePermission: []api.Permission{{
					EntityType:      string(entity.TypeClusterLink),
					EntityReference: entity.ClusterLinkURL("foo").String(),
					Entitlement:     "can_create_cluster_links",
				}},
				requiredPermission: serverPermission("can_create_cluster_links"),
				want:               false,
			},
			{
				condition:           "Empty permissions should not match",
				effectivePermission: nil,
				requiredPermission:  serverPermission("can_create_cluster_links"),
				want:                false,
			},
		}

		for _, c := range cases {
			var err error
			got := tunnel.HasPermission(c.effectivePermission, c.requiredPermission)
			if got != c.want {
				err = fmt.Errorf("expected %v, got %v", c.want, got)
			}

			helpers.LogTestOutcome(t, c.condition, err)
		}
	}
}

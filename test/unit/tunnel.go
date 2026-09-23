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

func testTunnel_CheckPermissions() (testName string, testFunc func(t *testing.T)) {
	return "Tunnel CheckPermissions matches effective LXD permissions", func(t *testing.T) {
		cases := []struct {
			condition            string
			effectivePermissions []api.Permission
			requiredPermissions  []api.Permission
			wantErr              bool
		}{
			{
				condition:            "Exact entitlement on the server entity should match",
				effectivePermissions: []api.Permission{serverPermission("can_create_cluster_links")},
				requiredPermissions:  []api.Permission{serverPermission("can_create_cluster_links")},
				wantErr:              false,
			},
			{
				condition:            "Admin entitlement on the server entity should imply any entitlement",
				effectivePermissions: []api.Permission{serverPermission("admin")},
				requiredPermissions:  []api.Permission{serverPermission("can_create_cluster_links")},
				wantErr:              false,
			},
			{
				condition:            "Admin entitlement on the server entity should imply permissions on other entities",
				effectivePermissions: []api.Permission{serverPermission("admin")},
				requiredPermissions: []api.Permission{{
					EntityType:      string(entity.TypeInstance),
					EntityReference: "/1.0/instances/c1?project=default",
					Entitlement:     "can_view",
				}},
				wantErr: false,
			},
			{
				condition:            "All required entitlements present should match",
				effectivePermissions: []api.Permission{serverPermission("can_view"), serverPermission("can_create_cluster_links")},
				requiredPermissions:  []api.Permission{serverPermission("can_view"), serverPermission("can_create_cluster_links")},
				wantErr:              false,
			},
			{
				condition:            "One missing required entitlement should not match",
				effectivePermissions: []api.Permission{serverPermission("can_view")},
				requiredPermissions:  []api.Permission{serverPermission("can_view"), serverPermission("can_create_cluster_links")},
				wantErr:              true,
			},
			{
				condition:            "A different entitlement on the server entity should not match",
				effectivePermissions: []api.Permission{serverPermission("can_view")},
				requiredPermissions:  []api.Permission{serverPermission("can_create_cluster_links")},
				wantErr:              true,
			},
			{
				condition: "The same entitlement on another entity should not match",
				effectivePermissions: []api.Permission{{
					EntityType:      string(entity.TypeClusterLink),
					EntityReference: entity.ClusterLinkURL("foo").String(),
					Entitlement:     "can_create_cluster_links",
				}},
				requiredPermissions: []api.Permission{serverPermission("can_create_cluster_links")},
				wantErr:             true,
			},
			{
				condition:            "Empty effective permissions should not match",
				effectivePermissions: nil,
				requiredPermissions:  []api.Permission{serverPermission("can_create_cluster_links")},
				wantErr:              true,
			},
			{
				condition:            "No required permissions should match",
				effectivePermissions: nil,
				requiredPermissions:  nil,
				wantErr:              false,
			},
		}

		for _, c := range cases {
			var err error
			gotErr := tunnel.CheckPermissions(c.effectivePermissions, c.requiredPermissions)
			if (gotErr != nil) != c.wantErr {
				err = fmt.Errorf("expected error %v, got %v", c.wantErr, gotErr)
			}

			helpers.LogTestOutcome(t, c.condition, err)
		}
	}
}

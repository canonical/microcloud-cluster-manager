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

func testClusterLink_RequiredPermissions_UnsupportedType() (testName string, testFunc func(t *testing.T)) {
	return "ClusterLink RequiredPermissions rejects unsupported link types", func(t *testing.T) {
		condition := "Unknown or missing link types should not be supported"

		var err error
		for _, linkType := range []string{"", "bogus"} {
			_, ok := clusterLink.RequirementsFor(linkType)
			if ok {
				err = fmt.Errorf("link type %q should not be supported", linkType)
				break
			}
		}

		helpers.LogTestOutcome(t, condition, err)
	}
}

func testClusterLink_RequirementsFor_SupportedTypes() (testName string, testFunc func(t *testing.T)) {
	return "ClusterLink RequirementsFor returns per-type requirements", func(t *testing.T) {
		condition := "Each supported link type should require the permissions of its creation flow"

		expected := map[string]struct{ source, target, sourceExtensions, targetExtensions []string }{
			"bidirectional": {
				[]string{"can_create_cluster_links", "can_create_identities"}, []string{"can_create_cluster_links", "can_create_identities"},
				[]string{"cluster_links"}, []string{"cluster_links"},
			},
			"unidirectional": {
				[]string{"can_create_cluster_links"}, []string{"can_create_cluster_links", "can_create_identities"},
				[]string{"cluster_links", "cluster_links_unidirectional"}, []string{"cluster_links", "cluster_links_unidirectional"},
			},
			"public": {
				[]string{"can_create_cluster_links"}, nil,
				[]string{"cluster_links", "cluster_links_public"}, nil,
			},
		}

		entitlements := func(permissions []api.Permission) []string {
			var result []string
			for _, permission := range permissions {
				result = append(result, permission.Entitlement)
			}

			return result
		}

		var err error
		for linkType, want := range expected {
			requirements, ok := clusterLink.RequirementsFor(linkType)
			if !ok {
				err = fmt.Errorf("link type %q should be supported", linkType)
				break
			}

			if !slices.Equal(entitlements(requirements.Source.Permissions), want.source) || !slices.Equal(entitlements(requirements.Target.Permissions), want.target) {
				err = fmt.Errorf("link type %q has unexpected permissions: source %v, target %v", linkType, entitlements(requirements.Source.Permissions), entitlements(requirements.Target.Permissions))
				break
			}
			if !slices.Equal(requirements.Source.APIExtensions, want.sourceExtensions) || !slices.Equal(requirements.Target.APIExtensions, want.targetExtensions) {
				err = fmt.Errorf("link type %q has unexpected API extensions: %+v", linkType, requirements)
				break
			}
			for _, permission := range append(slices.Clone(requirements.Source.Permissions), requirements.Target.Permissions...) {
				if permission.EntityType != string(entity.TypeServer) || permission.EntityReference != entity.ServerURL().String() {
					err = fmt.Errorf("link type %q has a non-server permission: %+v", linkType, permission)
					break
				}
			}
			if err != nil {
				break
			}

			requirements.Source.APIExtensions[0] = "changed"
			requirements.Source.Permissions[0].Entitlement = "changed"
			if len(requirements.Target.APIExtensions) > 0 {
				requirements.Target.APIExtensions[0] = "changed"
				requirements.Target.Permissions[0].Entitlement = "changed"
			}
			again, _ := clusterLink.RequirementsFor(linkType)
			if !slices.Equal(again.Source.APIExtensions, want.sourceExtensions) || !slices.Equal(again.Target.APIExtensions, want.targetExtensions) ||
				!slices.Equal(entitlements(again.Source.Permissions), want.source) || !slices.Equal(entitlements(again.Target.Permissions), want.target) {
				err = fmt.Errorf("link type %q requirements were mutated by a caller", linkType)
				break
			}
		}

		helpers.LogTestOutcome(t, condition, err)
	}
}

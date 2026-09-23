// Package cluster_link holds the cluster link creation rules shared by the API handlers:
// which LXD API extensions and entitlements each link type requires on the source and
// target clusters. Entitlements are evaluated via LXD's own effective-permissions API,
// not by reimplementing its authorization model.
package cluster_link

import (
	"github.com/canonical/lxd/lxd/auth"
	"github.com/canonical/lxd/shared/api"
	"github.com/canonical/lxd/shared/entity"
)

// ClusterLinksAPIExtension is the LXD API extension that adds cluster link support.
const ClusterLinksAPIExtension = "cluster_links"

// requiredEntitlements maps a cluster link type to the entitlements its creation flow
// requires on the source and target clusters. An empty list means health/reachability
// only for that side. Supporting a new link type only requires a new entry here.
var requiredEntitlements = map[string]struct {
	source []auth.Entitlement
	target []auth.Entitlement
}{
	"bidirectional": {
		source: []auth.Entitlement{auth.EntitlementCanCreateClusterLinks, auth.EntitlementCanCreateIdentities},
		target: []auth.Entitlement{auth.EntitlementCanCreateClusterLinks, auth.EntitlementCanCreateIdentities},
	},
}

// RequiredPermissions returns the server permissions the given cluster link type requires on
// the source and target clusters. ok is false if the link type is not supported.
func RequiredPermissions(linkType string) (source []api.Permission, target []api.Permission, ok bool) {
	required, ok := requiredEntitlements[linkType]
	if !ok {
		return nil, nil, false
	}

	toPermissions := func(entitlements []auth.Entitlement) []api.Permission {
		permissions := make([]api.Permission, 0, len(entitlements))
		for _, entitlement := range entitlements {
			permissions = append(permissions, api.Permission{
				EntityType:      string(entity.TypeServer),
				EntityReference: entity.ServerURL().String(),
				Entitlement:     string(entitlement),
			})
		}

		return permissions
	}

	return toPermissions(required.source), toPermissions(required.target), true
}

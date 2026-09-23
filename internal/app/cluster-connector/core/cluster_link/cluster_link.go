// Package cluster_link holds the cluster link creation rules shared by the API handlers:
// which LXD API extensions and entitlements each link type requires on the source and
// target clusters. Entitlements are evaluated via LXD's own effective-permissions API,
// not by reimplementing its authorization model.
package cluster_link

import (
	"github.com/canonical/lxd/shared/api"
	"github.com/canonical/lxd/shared/entity"
)

// ClusterLinksAPIExtension is the LXD API extension that adds cluster link support.
const ClusterLinksAPIExtension = "cluster_links"

// AccessManagementAPIExtension is the LXD API extension that adds the effective-permissions
// endpoint used to evaluate entitlements.
const AccessManagementAPIExtension = "access_management"

// Entitlement is an entitlement a cluster link creation flow requires on the server
// entity of a cluster, e.g. creating cluster links or identities.
type Entitlement string

const (
	// EntitlementCreateClusterLink allows creating cluster links on a cluster.
	EntitlementCreateClusterLink Entitlement = "can_create_cluster_links"
	// EntitlementCreateIdentity allows creating identities on a cluster.
	EntitlementCreateIdentity Entitlement = "can_create_identities"
	// EntitlementAdmin is the server-wide wildcard entitlement that implies all others.
	EntitlementAdmin Entitlement = "admin"
)

// requiredEntitlements maps a cluster link type to the entitlements its creation flow
// requires on the source and target clusters. An empty list means health/reachability
// only for that side. Supporting a new link type only requires a new entry here.
var requiredEntitlements = map[string]struct {
	source []Entitlement
	target []Entitlement
}{
	"bidirectional": {
		source: []Entitlement{EntitlementCreateClusterLink, EntitlementCreateIdentity},
		target: []Entitlement{EntitlementCreateClusterLink, EntitlementCreateIdentity},
	},
}

// RequiredEntitlements returns the entitlements the given cluster link type requires on the
// source and target clusters. ok is false if the link type is not supported.
func RequiredEntitlements(linkType string) (source []Entitlement, target []Entitlement, ok bool) {
	required, ok := requiredEntitlements[linkType]
	if !ok {
		return nil, nil, false
	}

	return required.source, required.target, true
}

// HasServerEntitlement reports whether the effective permissions reported by LXD include the
// given entitlement on the server entity, or the admin entitlement which implies it.
func HasServerEntitlement(effectivePermissions []api.Permission, entitlement Entitlement) bool {
	for _, p := range effectivePermissions {
		if p.EntityType != string(entity.TypeServer) || p.EntityReference != entity.ServerURL().String() {
			continue
		}

		if p.Entitlement == string(entitlement) || p.Entitlement == string(EntitlementAdmin) {
			return true
		}
	}

	return false
}

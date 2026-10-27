// Package cluster_link holds the cluster link creation rules shared by the API handlers:
// which LXD API extensions and entitlements each link type requires on the source and
// target clusters. Entitlements are evaluated via LXD's own effective-permissions API,
// not by reimplementing its authorization model.
package cluster_link

import (
	"slices"

	"github.com/canonical/lxd/lxd/auth"
	"github.com/canonical/lxd/shared/api"
	"github.com/canonical/lxd/shared/entity"
	"github.com/canonical/microcloud-cluster-manager/internal/app/cluster-connector/core/tunnel"
)

// Supported cluster link types. The public type is defined locally as the vendored LXD API does not expose it.
const (
	TypeBidirectional  = api.ClusterLinkTypeBidirectional
	TypeUnidirectional = api.ClusterLinkTypeUnidirectional
	TypePublic         = "public"
)

// ClusterLinksAPIExtension is the LXD API extension that adds cluster link support.
const ClusterLinksAPIExtension = "cluster_links"

const (
	// ClusterLinksUnidirectionalAPIExtension is the LXD API extension that adds unidirectional cluster links.
	ClusterLinksUnidirectionalAPIExtension = "cluster_links_unidirectional"
	// ClusterLinksPublicAPIExtension is the LXD API extension that adds public cluster links.
	ClusterLinksPublicAPIExtension = "cluster_links_public"
)

// ClusterLinkRequirements describes the capabilities needed on each cluster to create a link.
type ClusterLinkRequirements struct {
	Source tunnel.ClusterRequirements
	Target tunnel.ClusterRequirements
}

// requiredByType maps a cluster link type to what its creation flow requires on the source and
// target clusters. Supporting a new link type only requires a new entry here.
var requiredByType = map[string]ClusterLinkRequirements{
	TypeBidirectional: {
		Source: tunnel.ClusterRequirements{
			APIExtensions: []string{ClusterLinksAPIExtension},
			Permissions:   serverPermissions(auth.EntitlementCanCreateClusterLinks, auth.EntitlementCanCreateIdentities),
		},
		Target: tunnel.ClusterRequirements{
			APIExtensions: []string{ClusterLinksAPIExtension},
			Permissions:   serverPermissions(auth.EntitlementCanCreateClusterLinks, auth.EntitlementCanCreateIdentities),
		},
	},
	// The source (initiator) only stores the link, the target issues the pending identity and token.
	TypeUnidirectional: {
		Source: tunnel.ClusterRequirements{
			APIExtensions: []string{ClusterLinksAPIExtension, ClusterLinksUnidirectionalAPIExtension},
			Permissions:   serverPermissions(auth.EntitlementCanCreateClusterLinks),
		},
		Target: tunnel.ClusterRequirements{
			APIExtensions: []string{ClusterLinksAPIExtension, ClusterLinksUnidirectionalAPIExtension},
			Permissions:   serverPermissions(auth.EntitlementCanCreateClusterLinks, auth.EntitlementCanCreateIdentities),
		},
	},
	// Only the source is involved, the target is contacted by the source directly.
	TypePublic: {
		Source: tunnel.ClusterRequirements{
			APIExtensions: []string{ClusterLinksAPIExtension, ClusterLinksPublicAPIExtension},
			Permissions:   serverPermissions(auth.EntitlementCanCreateClusterLinks),
		},
	},
}

// RequirementsFor returns the requirements of the given cluster link type.
// ok is false if the link type is not supported.
func RequirementsFor(linkType string) (requirements ClusterLinkRequirements, ok bool) {
	required, ok := requiredByType[linkType]
	if !ok {
		return ClusterLinkRequirements{}, false
	}

	required.Source.APIExtensions = slices.Clone(required.Source.APIExtensions)
	required.Source.Permissions = slices.Clone(required.Source.Permissions)
	required.Target.APIExtensions = slices.Clone(required.Target.APIExtensions)
	required.Target.Permissions = slices.Clone(required.Target.Permissions)
	return required, true
}

func serverPermissions(entitlements ...auth.Entitlement) []api.Permission {
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

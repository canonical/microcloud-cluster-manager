// Package cluster_link resolves cluster links and their target resources.
package cluster_link

import (
	"strings"

	"github.com/canonical/lxd/shared"
	"github.com/canonical/lxd/shared/api"
)

// ClusterLinkTypeSupportsTargetCleanup reports whether a link type has a target-side resource to clean up.
func ClusterLinkTypeSupportsTargetCleanup(linkType string) bool {
	switch strings.ToLower(linkType) {
	case api.ClusterLinkTypeBidirectional, api.ClusterLinkTypeUnidirectional:
		return true
	default:
		return false
	}
}

// FindIdentity finds the cluster link identity for the certificate with the given fingerprint. LXD uses the
// certificate fingerprint as identifier of TLS identities; the fingerprint of the PEM encoded certificate of the
// identity is checked as well.
func FindIdentity(identities []api.Identity, fingerprint string) *api.Identity {
	if fingerprint == "" {
		return nil
	}

	for i, identity := range identities {
		if identity.Type != api.IdentityTypeCertificateClusterLink {
			continue
		}

		if strings.EqualFold(identity.Identifier, fingerprint) {
			return &identities[i]
		}

		if identity.TLSCertificate == "" {
			continue
		}

		certFingerprint, err := shared.CertFingerprintStr(identity.TLSCertificate)
		if err == nil && strings.EqualFold(certFingerprint, fingerprint) {
			return &identities[i]
		}
	}

	return nil
}

// FindReverseLinkByUUID finds the bidirectional link that points at the cluster with the given UUID.
// If no link matches, unresolved reports whether missing UUIDs prevented ruling out a reverse link.
func FindReverseLinkByUUID(links []api.ClusterLink, clusterUUID string) (reverseLink *api.ClusterLink, unresolved bool) {
	for i, link := range links {
		if !strings.EqualFold(link.Type, api.ClusterLinkTypeBidirectional) {
			continue
		}

		targetUUID := link.Config["volatile.uuid"]
		if targetUUID == "" || clusterUUID == "" {
			unresolved = true
			continue
		}
		if targetUUID == clusterUUID {
			return &links[i], false
		}
	}

	return nil, unresolved
}

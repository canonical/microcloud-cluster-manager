// Package cluster_link contains helpers to resolve both edges of a cluster link between remote clusters.
package cluster_link

import (
	"strings"

	"github.com/canonical/lxd/shared"
	"github.com/canonical/lxd/shared/api"
)

// LinkPointsAt reports whether the link points at the cluster with the given UUID, falling back to matching the
// link addresses against the LXD URL of that cluster.
func LinkPointsAt(link api.ClusterLink, clusterUUID string, lxdURL string) bool {
	if clusterUUID != "" && link.Config["volatile.uuid"] == clusterUUID {
		return true
	}

	if lxdURL == "" || link.Config["volatile.addresses"] == "" {
		return false
	}

	for _, address := range strings.Split(link.Config["volatile.addresses"], ",") {
		if address != "" && strings.HasSuffix(lxdURL, address) {
			return true
		}
	}

	return false
}

// FindReverseLink finds the bidirectional link pointing at the cluster with the given UUID, falling back to matching
// the link addresses against the LXD URL of that cluster. Links of other types are ignored, so an unrelated
// unidirectional or public link between the same clusters is never mistaken for the reverse edge.
func FindReverseLink(links []api.ClusterLink, clusterUUID string, lxdURL string) *api.ClusterLink {
	for i, link := range links {
		if !strings.EqualFold(link.Type, api.ClusterLinkTypeBidirectional) {
			continue
		}

		if LinkPointsAt(link, clusterUUID, lxdURL) {
			return &links[i]
		}
	}

	return nil
}

// FindLinkByName finds the link with the given name.
func FindLinkByName(links []api.ClusterLink, name string) *api.ClusterLink {
	for i, link := range links {
		if link.Name == name {
			return &links[i]
		}
	}

	return nil
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

package main

import (
	"errors"
	"strings"
	"testing"

	"github.com/canonical/lxd/shared"
	"github.com/canonical/lxd/shared/api"
	clusterLinkEdges "github.com/canonical/microcloud-cluster-manager/internal/app/management-api/core/cluster_link"
	"github.com/canonical/microcloud-cluster-manager/test/helpers"
)

func testClusterLinkEdges_FindIdentity() (testName string, testFunc func(t *testing.T)) {
	return "ClusterLink FindIdentity resolves the cluster link identity for a certificate fingerprint", func(t *testing.T) {
		condition := "Only cluster link identities whose identifier or certificate matches the fingerprint should match"

		var err error
		certPEM, _, genErr := shared.GenerateMemCert(false, shared.CertOptions{})
		if genErr != nil {
			helpers.LogTestOutcome(t, condition, genErr)
			return
		}

		fingerprint, fpErr := shared.CertFingerprintStr(string(certPEM))
		if fpErr != nil {
			helpers.LogTestOutcome(t, condition, fpErr)
			return
		}

		identities := []api.Identity{
			{Identifier: fingerprint, Type: api.IdentityTypeCertificateClient},
			{Identifier: "other", Type: api.IdentityTypeCertificateClusterLink},
			{Identifier: fingerprint, Type: api.IdentityTypeCertificateClusterLink},
		}

		byCertificate := []api.Identity{
			{Identifier: "other", Type: api.IdentityTypeCertificateClusterLink, TLSCertificate: "invalid"},
			{Identifier: "link", Type: api.IdentityTypeCertificateClusterLink, TLSCertificate: string(certPEM)},
		}

		identity := clusterLinkEdges.FindIdentity(identities, strings.ToUpper(fingerprint))
		if identity == nil || identity.Type != api.IdentityTypeCertificateClusterLink {
			err = errors.New("expected the cluster link identity with the fingerprint as identifier")
		} else if identity = clusterLinkEdges.FindIdentity(byCertificate, fingerprint); identity == nil || identity.Identifier != "link" {
			err = errors.New("expected the cluster link identity holding the certificate")
		} else if clusterLinkEdges.FindIdentity(identities, "") != nil {
			err = errors.New("an empty fingerprint should not match any identity")
		} else if clusterLinkEdges.FindIdentity(identities, "unknown") != nil {
			err = errors.New("an unknown fingerprint should not match any identity")
		}

		helpers.LogTestOutcome(t, condition, err)
	}
}

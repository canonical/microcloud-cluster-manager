package main

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/canonical/lxd/shared"
	"github.com/canonical/lxd/shared/api"
	clusterLinkEdges "github.com/canonical/microcloud-cluster-manager/internal/app/management-api/core/cluster_link"
	"github.com/canonical/microcloud-cluster-manager/test/helpers"
)

func testClusterLinkEdges_FindReverseLink() (testName string, testFunc func(t *testing.T)) {
	return "ClusterLink FindReverseLink resolves the link pointing back at a cluster", func(t *testing.T) {
		condition := "The reverse link should be bidirectional, match by UUID, fall back to addresses, and ignore links without either"

		links := []api.ClusterLink{
			{Name: "no-config", Type: api.ClusterLinkTypeBidirectional},
			{Name: "unidirectional", Type: api.ClusterLinkTypeUnidirectional, Config: map[string]string{"volatile.uuid": "source-uuid", "volatile.addresses": "10.0.0.1:8443"}},
			{Name: "other", Type: api.ClusterLinkTypeBidirectional, Config: map[string]string{"volatile.uuid": "other-uuid", "volatile.addresses": "10.0.0.9:8443"}},
			{Name: "by-address", Type: api.ClusterLinkTypeBidirectional, Config: map[string]string{"volatile.addresses": "10.0.0.8:8443,10.0.0.1:8443"}},
			{Name: "by-uuid", Type: api.ClusterLinkTypeBidirectional, Config: map[string]string{"volatile.uuid": "source-uuid"}},
		}

		var err error
		cases := []struct {
			uuid, lxdURL, expected string
		}{
			{"source-uuid", "", "by-uuid"},
			{"", "https://10.0.0.1:8443", "by-address"},
			{"unknown-uuid", "https://10.0.0.1:8443", "by-address"},
			{"", "", ""},
			{"unknown-uuid", "https://10.0.0.2:8443", ""},
		}

		for _, c := range cases {
			link := clusterLinkEdges.FindReverseLink(links, c.uuid, c.lxdURL)
			name := ""
			if link != nil {
				name = link.Name
			}

			if name != c.expected {
				err = fmt.Errorf("uuid %q and url %q matched link %q, expected %q", c.uuid, c.lxdURL, name, c.expected)
				break
			}
		}

		helpers.LogTestOutcome(t, condition, err)
	}
}

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

func testClusterLinkEdges_LinkPointsAt() (testName string, testFunc func(t *testing.T)) {
	return "ClusterLink LinkPointsAt checks whether a link points at a cluster", func(t *testing.T) {
		condition := "A link should point at a cluster matching its UUID or one of its addresses, and at no other cluster"

		link := api.ClusterLink{Name: "link", Config: map[string]string{"volatile.uuid": "target-uuid", "volatile.addresses": "10.0.0.1:8443"}}

		var err error
		cases := []struct {
			uuid, lxdURL string
			expected     bool
		}{
			{"target-uuid", "", true},
			{"", "https://10.0.0.1:8443", true},
			{"other-uuid", "https://10.0.0.2:8443", false},
			{"", "", false},
		}

		for _, c := range cases {
			if clusterLinkEdges.LinkPointsAt(link, c.uuid, c.lxdURL) != c.expected {
				err = fmt.Errorf("uuid %q and url %q: expected %t", c.uuid, c.lxdURL, c.expected)
				break
			}
		}

		if err == nil && clusterLinkEdges.LinkPointsAt(api.ClusterLink{Name: "no-config"}, "target-uuid", "https://10.0.0.1:8443") {
			err = errors.New("a link without config should not point at any cluster")
		}

		helpers.LogTestOutcome(t, condition, err)
	}
}

func testClusterLinkEdges_FindLinkByName() (testName string, testFunc func(t *testing.T)) {
	return "ClusterLink FindLinkByName resolves a link by its name", func(t *testing.T) {
		condition := "Only the link with the given name should match"

		links := []api.ClusterLink{{Name: "a"}, {Name: "b"}}

		var err error
		link := clusterLinkEdges.FindLinkByName(links, "b")
		if link == nil || link.Name != "b" {
			err = errors.New("expected link b")
		} else if clusterLinkEdges.FindLinkByName(links, "c") != nil {
			err = errors.New("an unknown name should not match any link")
		}

		helpers.LogTestOutcome(t, condition, err)
	}
}

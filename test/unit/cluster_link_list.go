package main

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/canonical/lxd/shared/api"
	clusterLinks "github.com/canonical/microcloud-cluster-manager/internal/app/management-api/core/cluster_link"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/database/store"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/request"
	"github.com/canonical/microcloud-cluster-manager/test/helpers"
)

func testClusterLinks_ParseListScope() (testName string, testFunc func(t *testing.T)) {
	return "ClusterLinks ParseListScope requires exactly one valid scope", func(t *testing.T) {
		condition := "A single non-empty cluster or all-clusters=true should be accepted, anything else should be a bad request"

		valid := []struct {
			query    string
			expected request.ListScope
		}{
			{"cluster=alpha", request.ListScope{Cluster: "alpha"}},
			{"cluster=a%20b", request.ListScope{Cluster: "a b"}},
			{"all-clusters=true", request.ListScope{AllClusters: true}},
			{"all-clusters=1", request.ListScope{AllClusters: true}},
		}

		invalid := []string{
			"",
			"other=alpha",
			"cluster=",
			"cluster=%20",
			"cluster=alpha&cluster=beta",
			"all-clusters=false",
			"all-clusters=",
			"all-clusters=yes",
			"all-clusters=true&all-clusters=true",
			"cluster=alpha&all-clusters=true",
			"cluster=%zz",
		}

		var err error
		for _, c := range valid {
			got, parseErr := request.ParseListScope(c.query)
			if parseErr != nil {
				err = fmt.Errorf("query %q failed: %w", c.query, parseErr)
				break
			}

			if got != c.expected {
				err = fmt.Errorf("query %q parsed to %v, expected %v", c.query, got, c.expected)
				break
			}
		}

		for _, query := range invalid {
			if err != nil {
				break
			}

			_, parseErr := request.ParseListScope(query)
			if !api.StatusErrorCheck(parseErr, http.StatusBadRequest) {
				err = fmt.Errorf("query %q returned %v, expected a bad request", query, parseErr)
			}
		}

		helpers.LogTestOutcome(t, condition, err)
	}
}

func testClusterLinks_ParseSourceCluster() (testName string, testFunc func(t *testing.T)) {
	return "ClusterLinks ParseSourceCluster requires exactly one source cluster", func(t *testing.T) {
		condition := "A single non-empty cluster should be accepted, anything else should be a bad request"

		valid := []struct {
			query    string
			expected string
		}{
			{"cluster=alpha", "alpha"},
			{"cluster=alpha&cleanupTarget=true", "alpha"},
		}

		invalid := []string{
			"",
			"cluster=",
			"cluster=%20",
			"cluster=alpha&cluster=beta",
			"all-clusters=true",
			"cluster=alpha&all-clusters=true",
			"cluster=%zz",
		}

		var err error
		for _, c := range valid {
			got, parseErr := request.ParseSourceCluster(c.query)
			if parseErr != nil {
				err = fmt.Errorf("query %q failed: %w", c.query, parseErr)
				break
			}

			if got != c.expected {
				err = fmt.Errorf("query %q parsed to %v, expected %v", c.query, got, c.expected)
				break
			}
		}

		for _, query := range invalid {
			if err != nil {
				break
			}

			_, parseErr := request.ParseSourceCluster(query)
			if !api.StatusErrorCheck(parseErr, http.StatusBadRequest) {
				err = fmt.Errorf("query %q returned %v, expected a bad request", query, parseErr)
			}
		}

		helpers.LogTestOutcome(t, condition, err)
	}
}

func testClusterLinks_TargetIndex() (testName string, testFunc func(t *testing.T)) {
	return "ClusterLinks TargetIndex resolves the target cluster of a link", func(t *testing.T) {
		condition := "Address matching should resolve only public links without a UUID, and the source should never resolve as the target"

		clusters := []store.RemoteClusterWithDetail{
			{Name: "other", ClusterUUID: "other-uuid", LXDURL: "https://10.0.0.1:8443"},
			{Name: "target", ClusterUUID: "target-uuid", LXDURL: "https://10.0.0.2:8443"},
			{Name: "without-uuid", LXDURL: "https://10.0.0.3:8443"},
			{Name: "ipv6", LXDURL: "https://[FD42::1]:8443"},
			{Name: "default-port", LXDURL: "https://lxd.example.com"},
		}

		cases := []struct {
			name     string
			linkType string
			source   string
			config   map[string]string
			expected string
		}{
			{"matching uuid without addresses", api.ClusterLinkTypeBidirectional, "other", map[string]string{"volatile.uuid": "target-uuid"}, "target"},
			{"public uuid with a different cluster's address", api.ClusterLinkTypePublic, "", map[string]string{"volatile.uuid": "target-uuid", "volatile.addresses": "10.0.0.1:8443"}, "target"},
			{"public source uuid with a different cluster's address", api.ClusterLinkTypePublic, "target", map[string]string{"volatile.uuid": "target-uuid", "volatile.addresses": "10.0.0.1:8443"}, ""},
			{"public unknown uuid with a matching address", api.ClusterLinkTypePublic, "", map[string]string{"volatile.uuid": "unknown", "volatile.addresses": "10.0.0.1:8443"}, ""},
			{"missing bidirectional uuid", api.ClusterLinkTypeBidirectional, "", map[string]string{"volatile.addresses": "10.0.0.1:8443"}, ""},
			{"missing unidirectional uuid", api.ClusterLinkTypeUnidirectional, "", map[string]string{"volatile.addresses": "10.0.0.1:8443"}, ""},
			{"empty uuid on link and cluster", api.ClusterLinkTypeBidirectional, "", map[string]string{"volatile.uuid": "", "volatile.addresses": "10.0.0.3:8443"}, ""},
			{"public confirmed address", api.ClusterLinkTypePublic, "", map[string]string{"volatile.addresses": "192.0.2.1:8443, 10.0.0.1:8443", "volatile.pending_address": "10.0.0.2:8443"}, "other"},
			{"public pending address", api.ClusterLinkTypePublic, "", map[string]string{"volatile.pending_address": " 10.0.0.1:8443 "}, "other"},
			{"public source address", api.ClusterLinkTypePublic, "other", map[string]string{"volatile.addresses": "10.0.0.1:8443"}, ""},
			{"public address without a managed match", api.ClusterLinkTypePublic, "", map[string]string{"volatile.addresses": "192.0.2.1:8443"}, ""},
			{"public address with default port", api.ClusterLinkTypePublic, "", map[string]string{"volatile.addresses": "LXD.example.com:8443"}, "default-port"},
			{"public IPv6 address", api.ClusterLinkTypePublic, "", map[string]string{"volatile.addresses": "[fd42::1]:8443"}, "ipv6"},
		}

		targets := clusterLinks.NewTargetIndex(clusters)
		var err error
		for _, c := range cases {
			link := api.ClusterLink{Name: c.name, Type: c.linkType, Config: c.config}
			target := clusterLinks.ResolveLinks(c.source, []api.ClusterLink{link}, targets)[0].TargetCluster
			name := ""
			if target != nil {
				name = *target
			}

			if name != c.expected {
				err = fmt.Errorf("case %q resolved to %q, expected %q", c.name, name, c.expected)
				break
			}
		}

		helpers.LogTestOutcome(t, condition, err)
	}
}

func testClusterLinks_FindReverseLinkByUUID() (testName string, testFunc func(t *testing.T)) {
	return "ClusterLinks distinguishes missing reverse links from unresolved ones", func(t *testing.T) {
		const sourceUUID = "source-uuid"
		pending := api.ClusterLink{Name: "pending", Type: api.ClusterLinkTypeBidirectional}
		reverse := api.ClusterLink{Name: "reverse", Type: api.ClusterLinkTypeBidirectional, Config: map[string]string{"volatile.uuid": sourceUUID}}
		unrelated := api.ClusterLink{Name: "unrelated", Type: api.ClusterLinkTypeBidirectional, Config: map[string]string{"volatile.uuid": "other-uuid"}}
		public := api.ClusterLink{Name: "public", Type: api.ClusterLinkTypePublic}

		cases := []struct {
			name       string
			links      []api.ClusterLink
			sourceUUID string
			wantName   string
			unresolved bool
		}{
			{"a matching UUID takes precedence", []api.ClusterLink{pending, reverse}, sourceUUID, reverse.Name, false},
			{"another UUID rules out a reverse link", []api.ClusterLink{unrelated}, sourceUUID, "", false},
			{"a missing target UUID prevents verification", []api.ClusterLink{pending, unrelated}, sourceUUID, "", true},
			{"a missing source UUID prevents verification", []api.ClusterLink{unrelated}, "", "", true},
			{"public links cannot be reverse links", []api.ClusterLink{public}, "", "", false},
		}

		for _, c := range cases {
			link, unresolved := clusterLinks.FindReverseLinkByUUID(c.links, c.sourceUUID)
			name := ""
			if link != nil {
				name = link.Name
			}
			var err error
			if name != c.wantName || unresolved != c.unresolved {
				err = fmt.Errorf("got link %q with unresolved=%t, want link %q with unresolved=%t", name, unresolved, c.wantName, c.unresolved)
			}
			helpers.LogTestOutcome(t, c.name, err)
		}
	}
}

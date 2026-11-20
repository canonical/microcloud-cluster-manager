package main

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/canonical/lxd/shared/api"
	clusterLinks "github.com/canonical/microcloud-cluster-manager/internal/app/management-api/core/cluster_link"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/database/store"
	"github.com/canonical/microcloud-cluster-manager/test/helpers"
)

func testClusterLinks_ParseListScope() (testName string, testFunc func(t *testing.T)) {
	return "ClusterLinks ParseListScope requires exactly one valid scope", func(t *testing.T) {
		condition := "A single non-empty cluster or all-clusters=true should be accepted, anything else should be a bad request"

		valid := []struct {
			query    string
			expected clusterLinks.ListScope
		}{
			{"cluster=alpha", clusterLinks.ListScope{Cluster: "alpha"}},
			{"cluster=a%20b", clusterLinks.ListScope{Cluster: "a b"}},
			{"all-clusters=true", clusterLinks.ListScope{AllClusters: true}},
			{"all-clusters=1", clusterLinks.ListScope{AllClusters: true}},
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

		err := checkValidQueries(valid, clusterLinks.ParseListScope)
		if err == nil {
			err = checkBadRequests(invalid, func(query string) error {
				_, err := clusterLinks.ParseListScope(query)
				return err
			})
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
			{"cluster=alpha&removeReverseEdge=true", "alpha"},
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

		err := checkValidQueries(valid, clusterLinks.ParseSourceCluster)
		if err == nil {
			err = checkBadRequests(invalid, func(query string) error {
				_, err := clusterLinks.ParseSourceCluster(query)
				return err
			})
		}

		helpers.LogTestOutcome(t, condition, err)
	}
}

func checkValidQueries[T comparable](cases []struct {
	query    string
	expected T
}, parse func(string) (T, error)) error {
	for _, c := range cases {
		got, err := parse(c.query)
		if err != nil {
			return fmt.Errorf("query %q failed: %w", c.query, err)
		}

		if got != c.expected {
			return fmt.Errorf("query %q parsed to %v, expected %v", c.query, got, c.expected)
		}
	}

	return nil
}

func checkBadRequests(queries []string, parse func(string) error) error {
	for _, query := range queries {
		err := parse(query)
		if !api.StatusErrorCheck(err, http.StatusBadRequest) {
			return fmt.Errorf("query %q returned %v, expected a bad request", query, err)
		}
	}

	return nil
}

func testClusterLinks_TargetIndex() (testName string, testFunc func(t *testing.T)) {
	return "ClusterLinks TargetIndex resolves the target cluster of a link", func(t *testing.T) {
		condition := "UUID matches should win over address matches, pending addresses should be used for links without addresses, the source cluster should never be the target, and unknown addresses should not resolve"

		clusters := []store.RemoteClusterWithDetail{
			{Name: "by-address", ClusterUUID: "address-uuid", LXDURL: "https://10.0.0.1:8443"},
			{Name: "by-uuid", ClusterUUID: "target-uuid", LXDURL: "https://10.0.0.2:8443"},
			{Name: "ipv6", LXDURL: "https://[FD42::1]:8443"},
			{Name: "default-port", LXDURL: "https://lxd.example.com"},
			{Name: "explicit-port", LXDURL: "https://lxd.example.com:443"},
			{Name: "no-url"},
		}

		cases := []struct {
			name     string
			source   string
			config   map[string]string
			expected string
		}{
			{"uuid over address", "", map[string]string{"volatile.uuid": "target-uuid", "volatile.addresses": "10.0.0.1:8443"}, "by-uuid"},
			{"source uuid falls back to address", "by-uuid", map[string]string{"volatile.uuid": "target-uuid", "volatile.addresses": "10.0.0.1:8443"}, "by-address"},
			{"source address", "by-address", map[string]string{"volatile.addresses": "10.0.0.1:8443"}, ""},
			{"unknown uuid falls back to address", "", map[string]string{"volatile.uuid": "unknown", "volatile.addresses": "10.0.0.1:8443"}, "by-address"},
			{"alternate address", "", map[string]string{"volatile.addresses": "192.0.2.1:8443, 10.0.0.1:8443"}, "by-address"},
			{"pending address", "", map[string]string{"volatile.pending_address": "10.0.0.1:8443"}, "by-address"},
			{"addresses over pending address", "", map[string]string{"volatile.addresses": "192.0.2.1:8443", "volatile.pending_address": "10.0.0.1:8443"}, ""},
			{"ipv6 case insensitive", "", map[string]string{"volatile.addresses": "[fd42::1]:8443"}, "ipv6"},
			{"default lxd port", "", map[string]string{"volatile.addresses": "lxd.example.com:8443"}, "default-port"},
			{"explicit https port", "", map[string]string{"volatile.addresses": "lxd.example.com:443"}, "explicit-port"},
			{"other port", "", map[string]string{"volatile.addresses": "10.0.0.1:9443"}, ""},
			{"address suffix", "", map[string]string{"volatile.addresses": "0.0.1:8443"}, ""},
			{"unknown address", "", map[string]string{"volatile.addresses": "192.0.2.1:8443"}, ""},
			{"no target config", "", nil, ""},
		}

		targets := clusterLinks.NewTargetIndex(clusters)
		var err error
		for _, c := range cases {
			target := targets.Resolve(api.ClusterLink{Name: c.name, Config: c.config}, c.source)
			name := ""
			if target != nil {
				name = target.Name
			}

			if name != c.expected {
				err = fmt.Errorf("case %q resolved to %q, expected %q", c.name, name, c.expected)
				break
			}
		}

		helpers.LogTestOutcome(t, condition, err)
	}
}

func testClusterLinks_ListDeadline() (testName string, testFunc func(t *testing.T)) {
	return "ClusterLinks ListDeadline leaves time to write the response", func(t *testing.T) {
		condition := "The deadline should be capped at 15s and end before the server write timeout"

		cases := []struct {
			writeTimeout time.Duration
			expected     time.Duration
		}{
			{10 * time.Second, 9 * time.Second},
			{4 * time.Second, 3 * time.Second},
			{1 * time.Second, 500 * time.Millisecond},
			{30 * time.Second, 15 * time.Second},
			{0, 15 * time.Second},
		}

		var err error
		for _, c := range cases {
			got := clusterLinks.ListDeadline(c.writeTimeout)
			if got != c.expected {
				err = fmt.Errorf("write timeout %v gave deadline %v, expected %v", c.writeTimeout, got, c.expected)
				break
			}
		}

		helpers.LogTestOutcome(t, condition, err)
	}
}

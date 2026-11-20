package main

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/canonical/microcloud-cluster-manager/test/helpers"
	"github.com/getkin/kin-openapi/routers"
)

// clusterLinkEndpoint is a cluster link endpoint scoped to a single managed cluster with the cluster query parameter.
type clusterLinkEndpoint struct {
	method string
	// path is the request path, without the cluster query parameter.
	path  string
	query string
	body  any
}

var clusterLinkEndpoints = []clusterLinkEndpoint{
	{http.MethodGet, "/1.0/cluster-links", "", nil},
	{http.MethodPost, "/1.0/cluster-links", "", map[string]any{"name": "link", "type": "bidirectional", "target_cluster": "other"}},
	{http.MethodPut, "/1.0/cluster-links/link", "", map[string]any{"description": "updated"}},
	{http.MethodDelete, "/1.0/cluster-links/link", "&cleanupTarget=true", nil},
}

func clusterLinkPath(endpoint clusterLinkEndpoint, clusterName string) string {
	return endpoint.path + "?cluster=" + url.QueryEscape(clusterName) + endpoint.query
}

func testListAllClusterLinksSuccess(env *helpers.Environment, router routers.Router) (testName string, testFunc func(t *testing.T)) {
	return "testListAllClusterLinksSuccess GET /1.0/cluster-links?all-clusters=true 200", func(t *testing.T) {
		condition := "200: listing the cluster links of all clusters returns the links and per-cluster errors."
		err := helpers.EnforceSuccessSchema(env, router, http.MethodGet, "/1.0/cluster-links?all-clusters=true", nil, http.StatusOK)
		helpers.LogTestOutcome(t, condition, err)
	}
}

func testListClusterLinksWithoutScope(env *helpers.Environment, router routers.Router) (testName string, testFunc func(t *testing.T)) {
	return "testListClusterLinksWithoutScope GET /1.0/cluster-links 400", func(t *testing.T) {
		condition := "400: listing cluster links without a cluster or all-clusters returns error response."
		err := helpers.EnforceSuccessSchema(env, router, http.MethodGet, "/1.0/cluster-links", nil, http.StatusBadRequest)
		helpers.LogTestOutcome(t, condition, err)
	}
}

func testClusterLinkEndpointsNotFound(env *helpers.Environment, router routers.Router) (testName string, testFunc func(t *testing.T)) {
	return "testClusterLinkEndpointsNotFound /1.0/cluster-links...?cluster={name} 404", func(t *testing.T) {
		for _, endpoint := range clusterLinkEndpoints {
			condition := fmt.Sprintf("404: %s %s for a non-existent cluster returns error response.", endpoint.method, endpoint.path)
			err := helpers.EnforceSuccessSchema(env, router, endpoint.method, clusterLinkPath(endpoint, "definitely-does-not-exist-99999"), endpoint.body, http.StatusNotFound)
			helpers.LogTestOutcome(t, condition, err)
		}
	}
}

func testClusterLinkEndpointsNoTunnel(env *helpers.Environment, router routers.Router) (testName string, testFunc func(t *testing.T)) {
	return "testClusterLinkEndpointsNoTunnel /1.0/cluster-links...?cluster={name} 400", func(t *testing.T) {
		clusterName, err := helpers.CreateRandomCluster(env)
		if err != nil {
			t.Fatalf("Failed to create cluster for cluster link endpoints test: %v", err)
		}

		t.Cleanup(func() {
			err := helpers.DeleteRemoteCluster(env, clusterName)
			if err != nil {
				t.Logf("Failed to delete cluster %q: %v", clusterName, err)
			}
		})

		for _, endpoint := range clusterLinkEndpoints {
			condition := fmt.Sprintf("400: %s %s for a cluster without an active tunnel returns error response.", endpoint.method, endpoint.path)
			err := helpers.EnforceSuccessSchema(env, router, endpoint.method, clusterLinkPath(endpoint, clusterName), endpoint.body, http.StatusBadRequest)
			helpers.LogTestOutcome(t, condition, err)
		}
	}
}

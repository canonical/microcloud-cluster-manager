package main

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/canonical/microcloud-cluster-manager/test/helpers"
	"github.com/getkin/kin-openapi/routers"
)

// forwardedEndpoint is a remote cluster endpoint that is forwarded to the cluster connector.
type forwardedEndpoint struct {
	method string
	// suffix is appended to /1.0/remote-cluster/{remoteClusterName}.
	suffix string
	body   any
}

var forwardedEndpoints = []forwardedEndpoint{
	{http.MethodGet, "/auth-groups", nil},
	{http.MethodGet, "/auth/identities", nil},
	{http.MethodGet, "/cluster-links", nil},
	{http.MethodPost, "/cluster-links", map[string]any{"name": "link", "type": "bidirectional", "target_cluster": "other"}},
	{http.MethodPut, "/cluster-links/link", map[string]any{"description": "updated"}},
	{http.MethodDelete, "/cluster-links/link?removeReverseEdge=true", nil},
	{http.MethodGet, "/replicators?project=default", nil},
	{http.MethodPost, "/replicators?project=default", map[string]any{"name": "replicator"}},
	{http.MethodGet, "/replicators/replicator?project=default", nil},
	{http.MethodPut, "/replicators/replicator?project=default", map[string]any{"description": "updated"}},
	{http.MethodDelete, "/replicators/replicator?project=default", nil},
	{http.MethodGet, "/replicators/replicator/state?project=default", nil},
	{http.MethodPut, "/replicators/replicator/state?project=default", map[string]any{"action": "start"}},
	{http.MethodGet, "/projects", nil},
	{http.MethodPost, "/projects", map[string]any{"name": "project"}},
	{http.MethodGet, "/projects/project", nil},
	{http.MethodPut, "/projects/project", map[string]any{"description": "updated"}},
	{http.MethodGet, "/projects/project/state", nil},
	{http.MethodPut, "/projects/project/state?force=true", map[string]any{"replica_mode": "leader"}},
	{http.MethodGet, "/operations/00000000-0000-0000-0000-000000000000", nil},
}

func testForwardedRemoteClusterEndpointsNotFound(env *helpers.Environment, router routers.Router) (testName string, testFunc func(t *testing.T)) {
	return "testForwardedRemoteClusterEndpointsNotFound /1.0/remote-cluster/{remoteClusterName}/... 404", func(t *testing.T) {
		for _, endpoint := range forwardedEndpoints {
			path := "/1.0/remote-cluster/definitely-does-not-exist-99999" + endpoint.suffix
			condition := fmt.Sprintf("404: %s %s for a non-existent cluster returns error response.", endpoint.method, endpoint.suffix)
			err := helpers.EnforceSuccessSchema(env, router, endpoint.method, path, endpoint.body, http.StatusNotFound)
			helpers.LogTestOutcome(t, condition, err)
		}
	}
}

func testForwardedRemoteClusterEndpointsNoTunnel(env *helpers.Environment, router routers.Router) (testName string, testFunc func(t *testing.T)) {
	return "testForwardedRemoteClusterEndpointsNoTunnel /1.0/remote-cluster/{remoteClusterName}/... 400", func(t *testing.T) {
		clusterName, err := helpers.CreateRandomCluster(env)
		if err != nil {
			t.Fatalf("Failed to create cluster for forwarded endpoints test: %v", err)
		}

		t.Cleanup(func() {
			err := helpers.DeleteRemoteCluster(env, clusterName)
			if err != nil {
				t.Logf("Failed to delete cluster %q: %v", clusterName, err)
			}
		})

		for _, endpoint := range forwardedEndpoints {
			path := "/1.0/remote-cluster/" + clusterName + endpoint.suffix
			condition := fmt.Sprintf("400: %s %s for a cluster without an active tunnel returns error response.", endpoint.method, endpoint.suffix)
			err := helpers.EnforceSuccessSchema(env, router, endpoint.method, path, endpoint.body, http.StatusBadRequest)
			helpers.LogTestOutcome(t, condition, err)
		}
	}
}

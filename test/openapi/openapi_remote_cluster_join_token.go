package main

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/canonical/microcloud-cluster-manager/internal/pkg/api/models/v1"
	"github.com/canonical/microcloud-cluster-manager/test/helpers"
	"github.com/getkin/kin-openapi/routers"
)

func testListRemoteClusterJoinTokensSuccess(env *helpers.Environment, router routers.Router) (testName string, testFunc func(t *testing.T)) {
	return "testListRemoteClusterJoinTokensSuccess GET /1.0/remote-cluster-join-token", func(t *testing.T) {
		condition := "200: authenticated request returns list of join tokens."
		err := helpers.TestOpenAPIOperation(t, env, router, http.MethodGet, "/1.0/remote-cluster-join-token", nil)
		helpers.LogTestOutcome(t, condition, err)
	}
}

func testCreateRemoteClusterJoinTokenSuccess(env *helpers.Environment, router routers.Router) (testName string, testFunc func(t *testing.T)) {
	return "testCreateRemoteClusterJoinTokenSuccess POST /1.0/remote-cluster-join-token 200", func(t *testing.T) {
		clusterName := "cluster-100"

		body := models.RemoteClusterTokenPost{
			ClusterName: clusterName,
			Description: "created by OpenAPI schema test",
			Expiry:      time.Now().Add(24 * time.Hour),
		}

		condition := "200: creating a join token with a valid payload returns token response."
		err := helpers.TestOpenAPIOperation(t, env, router, http.MethodPost, "/1.0/remote-cluster-join-token", body)
		helpers.LogTestOutcome(t, condition, err)
	}
}

func testCreateRemoteClusterJoinTokenBadRequest(env *helpers.Environment, router routers.Router) (testName string, testFunc func(t *testing.T)) {
	return "testCreateRemoteClusterJoinTokenBadRequest POST /1.0/remote-cluster-join-token 400", func(t *testing.T) {
		// Omit the required `cluster_name` field — the server must respond with 400.
		// Request validation is intentionally bypassed since the payload violates the schema.
		invalidBody := map[string]any{
			"description": "missing required cluster_name",
		}

		condition := "400: creating a token without the required cluster_name returns error response."
		err := helpers.TestOpenAPIOperationInvalid(t, env, router, http.MethodPost, "/1.0/remote-cluster-join-token", invalidBody)
		helpers.LogTestOutcome(t, condition, err)
	}
}

func testCreateRemoteClusterJoinTokenConflict(env *helpers.Environment, router routers.Router) (testName string, testFunc func(t *testing.T)) {
	return "testCreateRemoteClusterJoinTokenConflict POST /1.0/remote-cluster-join-token 409", func(t *testing.T) {
		clusterName := "cluster-01" // From seed data, already has a join token

		body := models.RemoteClusterTokenPost{
			ClusterName: clusterName,
			Description: "created by OpenAPI schema test",
			Expiry:      time.Now().Add(24 * time.Hour),
		}

		condition := "409: creating a join token for a cluster that already has one returns error response."
		err := helpers.TestOpenAPIOperation(t, env, router, http.MethodPost, "/1.0/remote-cluster-join-token", body)
		helpers.LogTestOutcome(t, condition, err)
	}
}

func testDeleteRemoteClusterJoinTokenSuccess(env *helpers.Environment, router routers.Router) (testName string, testFunc func(t *testing.T)) {
	return "testDeleteRemoteClusterJoinTokenSuccess DELETE /1.0/remote-cluster-join-token/{remoteClusterName} 200", func(t *testing.T) {
		clusterName := "cluster-03"
		condition := "200: deleting an existing join token returns empty response."
		err := helpers.TestOpenAPIOperation(t, env, router, http.MethodDelete, fmt.Sprintf("/1.0/remote-cluster-join-token/%s", clusterName), nil)
		helpers.LogTestOutcome(t, condition, err)
	}
}

func testDeleteRemoteClusterJoinTokenNotFound(env *helpers.Environment, router routers.Router) (testName string, testFunc func(t *testing.T)) {
	return "testDeleteRemoteClusterJoinTokenNotFound DELETE /1.0/remote-cluster-join-token/{remoteClusterName} 404", func(t *testing.T) {
		clusterName := "non-existent-cluster"
		condition := "404: deleting a non-existent join token returns error response."
		err := helpers.TestOpenAPIOperation(t, env, router, http.MethodDelete, fmt.Sprintf("/1.0/remote-cluster-join-token/%s", clusterName), nil)
		helpers.LogTestOutcome(t, condition, err)
	}
}

func testDeleteRemoteClusterJoinTokenInternalError(env *helpers.Environment, router routers.Router) (testName string, testFunc func(t *testing.T)) {
	return "testDeleteRemoteClusterJoinTokenInternalError DELETE /1.0/remote-cluster-join-token/{remoteClusterName} 500", func(t *testing.T) {
		malformedClusterName := "%g1" // A cluster name that is improperly escaped to trigger internal error in handler
		condition := "500: deleting a join token with a malformed cluster name returns error response."
		err := helpers.TestOpenAPIOperationInvalid(t, env, router, http.MethodDelete, fmt.Sprintf("/1.0/remote-cluster-join-token/%s", malformedClusterName), nil)
		helpers.LogTestOutcome(t, condition, err)
	}
}

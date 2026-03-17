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
		err := helpers.EnforceSuccessSchema(env, router, http.MethodGet, "/1.0/remote-cluster-join-token", nil, http.StatusOK)
		helpers.LogTestOutcome(t, condition, err)
	}
}

func testCreateRemoteClusterJoinTokenSuccess(env *helpers.Environment, router routers.Router) (testName string, testFunc func(t *testing.T)) {
	return "testCreateRemoteClusterJoinTokenSuccess POST /1.0/remote-cluster-join-token 200", func(t *testing.T) {
		clusterName := helpers.GetRandomName("openapitest-cluster")
		body := models.RemoteClusterTokenPost{
			ClusterName: clusterName,
			Description: "created by OpenAPI schema test",
			Expiry:      time.Now().Add(24 * time.Hour),
		}

		condition := "200: creating a join token with a valid payload returns token response."
		err := helpers.EnforceSuccessSchema(env, router, http.MethodPost, "/1.0/remote-cluster-join-token", body, http.StatusOK)
		helpers.LogTestOutcome(t, condition, err)
	}
}

func testCreateRemoteClusterJoinTokenBadRequest(env *helpers.Environment, router routers.Router) (testName string, testFunc func(t *testing.T)) {
	return "testCreateRemoteClusterJoinTokenBadRequest POST /1.0/remote-cluster-join-token 400", func(t *testing.T) {
		invalidBody := map[string]any{
			"description": "missing required cluster_name",
		}

		condition := "400: creating a token without the required cluster_name returns error response."
		err := helpers.EnforceErrorSchema(env, router, http.MethodPost, "/1.0/remote-cluster-join-token", invalidBody, http.StatusBadRequest)
		helpers.LogTestOutcome(t, condition, err)
	}
}

func testCreateRemoteClusterJoinTokenConflict(env *helpers.Environment, router routers.Router) (testName string, testFunc func(t *testing.T)) {
	return "testCreateRemoteClusterJoinTokenConflict POST /1.0/remote-cluster-join-token 409", func(t *testing.T) {
		clusterName, err := helpers.CreateRandomJoinToken(env)
		if err != nil {
			t.Fatalf("Failed to create join token for conflict test: %v", err)
		}

		body := models.RemoteClusterTokenPost{
			ClusterName: clusterName,
			Description: "created by OpenAPI schema test",
			Expiry:      time.Now().Add(24 * time.Hour),
		}

		condition := "409: creating a join token for a cluster that already has one returns error response."
		err = helpers.EnforceSuccessSchema(env, router, http.MethodPost, "/1.0/remote-cluster-join-token", body, http.StatusConflict)
		helpers.LogTestOutcome(t, condition, err)
	}
}

func testDeleteRemoteClusterJoinTokenSuccess(env *helpers.Environment, router routers.Router) (testName string, testFunc func(t *testing.T)) {
	return "testDeleteRemoteClusterJoinTokenSuccess DELETE /1.0/remote-cluster-join-token/{remoteClusterName} 200", func(t *testing.T) {
		clusterName, err := helpers.CreateRandomJoinToken(env)
		if err != nil {
			t.Fatalf("Failed to create join token for deletion test: %v", err)
		}

		condition := "200: deleting an existing join token returns empty response."
		err = helpers.EnforceSuccessSchema(env, router, http.MethodDelete, fmt.Sprintf("/1.0/remote-cluster-join-token/%s", clusterName), nil, http.StatusOK)
		helpers.LogTestOutcome(t, condition, err)
	}
}

func testDeleteRemoteClusterJoinTokenNotFound(env *helpers.Environment, router routers.Router) (testName string, testFunc func(t *testing.T)) {
	return "testDeleteRemoteClusterJoinTokenNotFound DELETE /1.0/remote-cluster-join-token/{remoteClusterName} 404", func(t *testing.T) {
		clusterName := "non-existent-cluster"
		condition := "404: deleting a non-existent join token returns error response."
		err := helpers.EnforceSuccessSchema(env, router, http.MethodDelete, fmt.Sprintf("/1.0/remote-cluster-join-token/%s", clusterName), nil, http.StatusNotFound)
		helpers.LogTestOutcome(t, condition, err)
	}
}

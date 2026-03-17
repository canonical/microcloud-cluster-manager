package main

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/canonical/microcloud-cluster-manager/internal/pkg/api/models/v1"
	"github.com/canonical/microcloud-cluster-manager/test/helpers"
	"github.com/getkin/kin-openapi/routers"
)

func testListRemoteClustersSuccess(env *helpers.Environment, router routers.Router) (testName string, testFunc func(t *testing.T)) {
	return "testListRemoteClustersSuccess GET /1.0/remote-cluster", func(t *testing.T) {
		condition := "200: authenticated request returns list of remote clusters."
		err := helpers.EnforceSuccessSchema(env, router, http.MethodGet, "/1.0/remote-cluster", nil, http.StatusOK)
		helpers.LogTestOutcome(t, condition, err)
	}
}

func testGetRemoteClusterSuccess(env *helpers.Environment, router routers.Router) (testName string, testFunc func(t *testing.T)) {
	return "testGetRemoteClusterSuccess GET /1.0/remote-cluster/{remoteClusterName} 200", func(t *testing.T) {
		clusterName, err := helpers.CreateRandomCluster(env)
		if err != nil {
			t.Fatalf("Failed to register cluster for get test: %v", err)
		}

		condition := "200: fetching an existing cluster returns metadata."
		err = helpers.EnforceSuccessSchema(env, router, http.MethodGet, fmt.Sprintf("/1.0/remote-cluster/%s", clusterName), nil, http.StatusOK)
		helpers.LogTestOutcome(t, condition, err)
	}
}

func testGetRemoteClusterNotFound(env *helpers.Environment, router routers.Router) (testName string, testFunc func(t *testing.T)) {
	return "testGetRemoteClusterNotFound GET /1.0/remote-cluster/{remoteClusterName} 404", func(t *testing.T) {
		condition := "404: fetching a non-existent cluster returns error response."
		err := helpers.EnforceSuccessSchema(env, router, http.MethodGet, "/1.0/remote-cluster/definitely-does-not-exist-99999", nil, http.StatusNotFound)
		helpers.LogTestOutcome(t, condition, err)
	}
}

func testPatchRemoteClusterSuccess(env *helpers.Environment, router routers.Router) (testName string, testFunc func(t *testing.T)) {
	return "testPatchRemoteClusterSuccess PATCH /1.0/remote-cluster/{remoteClusterName} 200", func(t *testing.T) {
		clusterName, err := helpers.CreateRandomCluster(env)
		if err != nil {
			t.Fatalf("Failed to create cluster for patch test: %v", err)
		}

		patch := models.RemoteClusterPatch{
			Status:          models.ACTIVE,
			Description:     "updated by OpenAPI schema test",
			DiskThreshold:   75,
			MemoryThreshold: 80,
		}

		condition := "200: patching an existing cluster with a valid payload returns empty response."
		err = helpers.EnforceSuccessSchema(env, router, http.MethodPatch, fmt.Sprintf("/1.0/remote-cluster/%s", clusterName), patch, http.StatusOK)
		helpers.LogTestOutcome(t, condition, err)
	}
}

func testPatchRemoteClusterNotFound(env *helpers.Environment, router routers.Router) (testName string, testFunc func(t *testing.T)) {
	return "testPatchRemoteClusterNotFound PATCH /1.0/remote-cluster/{remoteClusterName} 404", func(t *testing.T) {
		clusterName := "non-existent-cluster"
		patch := models.RemoteClusterPatch{
			Status:          models.ACTIVE,
			Description:     "updated by OpenAPI schema test",
			DiskThreshold:   75,
			MemoryThreshold: 80,
		}

		condition := "404: patching a non-existent cluster returns error response."
		err := helpers.EnforceSuccessSchema(env, router, http.MethodPatch, fmt.Sprintf("/1.0/remote-cluster/%s", clusterName), patch, http.StatusNotFound)
		helpers.LogTestOutcome(t, condition, err)
	}
}

func testPatchRemoteClusterBadRequest(env *helpers.Environment, router routers.Router) (testName string, testFunc func(t *testing.T)) {
	return "testPatchRemoteClusterBadRequest PATCH /1.0/remote-cluster/{remoteClusterName} 400", func(t *testing.T) {
		clusterName, err := helpers.CreateRandomCluster(env)
		if err != nil {
			t.Fatalf("Failed to create cluster for patch test: %v", err)
		}

		invalidPatch := map[string]any{
			"status": "DEFINITELY_INVALID_STATUS",
		}

		condition := "400: patching with an invalid status enum value returns error response."
		err = helpers.EnforceErrorSchema(env, router, http.MethodPatch, fmt.Sprintf("/1.0/remote-cluster/%s", clusterName), invalidPatch, http.StatusBadRequest)
		helpers.LogTestOutcome(t, condition, err)
	}
}

func testDeleteRemoteClusterSuccess(env *helpers.Environment, router routers.Router) (testName string, testFunc func(t *testing.T)) {
	return "testDeleteRemoteClusterSuccess DELETE /1.0/remote-cluster/{remoteClusterName} 200", func(t *testing.T) {
		clusterName, err := helpers.CreateRandomCluster(env)
		if err != nil {
			t.Fatalf("Failed to create cluster for deletion test: %v", err)
		}

		condition := "200: deleting an existing cluster returns empty response"
		err = helpers.EnforceSuccessSchema(env, router, http.MethodDelete, fmt.Sprintf("/1.0/remote-cluster/%s", clusterName), nil, http.StatusOK)
		helpers.LogTestOutcome(t, condition, err)
	}
}

func testDeleteRemoteClusterNotFound(env *helpers.Environment, router routers.Router) (testName string, testFunc func(t *testing.T)) {
	return "testDeleteRemoteClusterNotFound DELETE /1.0/remote-cluster/{remoteClusterName} 404", func(t *testing.T) {
		clusterName := "non-existent-cluster"
		condition := "404: deleting a non-existent cluster returns error response."
		err := helpers.EnforceSuccessSchema(env, router, http.MethodDelete, fmt.Sprintf("/1.0/remote-cluster/%s", clusterName), nil, http.StatusNotFound)
		helpers.LogTestOutcome(t, condition, err)
	}
}

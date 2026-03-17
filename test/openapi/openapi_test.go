package main

import (
	"context"
	"testing"

	"github.com/canonical/microcloud-cluster-manager/test/helpers"
	"github.com/canonical/microcloud-cluster-manager/test/types"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/routers/gorillamux"
)

var tests = []types.APISchemaTest{
	testRootEndpointSuccess,
	testStatusEndpointSuccess,
	testConfigurationEndpointSuccess,
	testListRemoteClustersSuccess,
	testGetRemoteClusterSuccess,
	testGetRemoteClusterNotFound,
	testPatchRemoteClusterSuccess,
	testPatchRemoteClusterNotFound,
	testPatchRemoteClusterBadRequest,
	testDeleteRemoteClusterSuccess,
	testDeleteRemoteClusterNotFound,
	testListRemoteClusterJoinTokensSuccess,
	testCreateRemoteClusterJoinTokenSuccess,
	testCreateRemoteClusterJoinTokenBadRequest,
	testCreateRemoteClusterJoinTokenConflict,
	testDeleteRemoteClusterJoinTokenSuccess,
	testDeleteRemoteClusterJoinTokenNotFound,
}

func TestOpenapiSchema(t *testing.T) {
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromFile("cluster-manager-api.yaml")
	if err != nil {
		t.Fatalf("Failed to load OpenAPI spec: %v", err)
	}

	err = doc.Validate(context.Background())
	if err != nil {
		t.Fatalf("OpenAPI spec validation failed: %v", err)
	}

	router, err := gorillamux.NewRouter(doc)

	if err != nil {
		t.Fatalf("Failed to create router from OpenAPI spec: %v", err)
	}

	env := initEnv(t)

	defer cleanup(t, env)

	for _, tt := range tests {
		testName, testFunc := tt(env, router)
		t.Run(testName, testFunc)
	}
}

func initEnv(t *testing.T) *helpers.Environment {
	env := helpers.NewEnv()
	err := env.Init()
	helpers.LogTestOutcome(t, "Initialize environment", err)
	return env
}

func cleanup(t *testing.T, env *helpers.Environment) {
	err := env.Cleanup()
	if err != nil {
		t.Fatalf("Failed to cleanup environment: %v", err)
	}
}

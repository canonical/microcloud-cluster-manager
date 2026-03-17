package main

import (
	"net/http"
	"testing"

	"github.com/canonical/microcloud-cluster-manager/test/helpers"
	"github.com/getkin/kin-openapi/routers"
)

func testConfigurationEndpointSuccess(env *helpers.Environment, router routers.Router) (testName string, testFunc func(t *testing.T)) {
	return "testConfigurationEndpointSuccess GET /1.0/configuration", func(t *testing.T) {
		condition := "200: authenticated request returns configuration."
		err := helpers.TestOpenAPIOperation(t, env, router, http.MethodGet, "/1.0/configuration", nil)
		helpers.LogTestOutcome(t, condition, err)
	}
}

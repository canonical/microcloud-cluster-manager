package main

import (
	"net/http"
	"testing"

	"github.com/canonical/microcloud-cluster-manager/test/helpers"
	"github.com/getkin/kin-openapi/routers"
)

func testRootEndpointSuccess(env *helpers.Environment, router routers.Router) (testName string, testFunc func(t *testing.T)) {
	return "testRootEndpointSuccess GET /1.0", func(t *testing.T) {
		condition := "200: authenticated request returns user metadata."
		err := helpers.EnforceSuccessSchema(env, router, http.MethodGet, "/1.0", nil, http.StatusOK)
		helpers.LogTestOutcome(t, condition, err)
	}
}

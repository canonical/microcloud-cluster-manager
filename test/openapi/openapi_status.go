package main

import (
	"net/http"
	"testing"

	"github.com/canonical/microcloud-cluster-manager/test/helpers"
	"github.com/getkin/kin-openapi/routers"
)

func testStatusEndpointSuccess(env *helpers.Environment, router routers.Router) (testName string, testFunc func(t *testing.T)) {
	return "testStatusEndpointSuccess GET /1.0/status", func(t *testing.T) {
		condition := "200: healthy service returns empty response."
		err := helpers.EnforceSuccessSchema(env, router, http.MethodGet, "/1.0/status", nil, http.StatusOK)
		helpers.LogTestOutcome(t, condition, err)
	}
}

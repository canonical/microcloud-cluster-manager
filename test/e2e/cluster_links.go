package main

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/canonical/lxd/shared/api"
	"github.com/canonical/microcloud-cluster-manager/test/helpers"
)

// testClusterLinksAPI verifies the scope validation and tunnel handling of the management API cluster
// links endpoints. The fake clusters used in e2e tests never open a tunnel, so their links cannot be loaded.
func testClusterLinksAPI(env *helpers.Environment) (testName string, testFunc func(t *testing.T)) {
	return "cluster links API validates scopes and skips clusters without a tunnel", func(t *testing.T) {
		clusterName := helpers.GetRandomName("cluster_links_e2e")

		defer func() {
			env.RemoveRemoteCluster(clusterName)
			env.RemoveRemoteClusterToken(clusterName)
		}()

		condition := "Should register a remote cluster"
		_, err := helpers.RegisterRemoteCluster(env, clusterName)
		helpers.LogTestOutcome(t, condition, err)
		if err != nil {
			return
		}

		certPublicKey, err := env.ManagementAPICert().PublicKeyX509()
		if err != nil {
			helpers.LogTestOutcome(t, "Should load the management API certificate", err)
			return
		}

		adminHeaders, err := env.ManagementAPILoginHeaders(certPublicKey)
		if err != nil {
			helpers.LogTestOutcome(t, "Should log in as admin", err)
			return
		}

		path := func(rawQuery string, segments ...string) *api.URL {
			u := api.NewURL().Scheme("https").Host(env.ManagementAPIHostPort()).Path(append([]string{"1.0", "cluster-links"}, segments...)...)
			u.RawQuery = rawQuery
			return u
		}

		clusterQuery := "cluster=" + clusterName
		cases := []struct {
			condition    string
			method       string
			url          *api.URL
			input        any
			expectStatus int
		}{
			{"Should reject a list without a scope", http.MethodGet, path(""), nil, http.StatusBadRequest},
			{"Should reject a create without a source cluster", http.MethodPost, path(""), map[string]any{}, http.StatusBadRequest},
			{"Should reject an update with a repeated cluster", http.MethodPut, path(clusterQuery+"&"+clusterQuery, "link"), map[string]any{}, http.StatusBadRequest},
			{"Should reject an unknown delete query parameter", http.MethodDelete, path(clusterQuery+"&cleanupTargte=true", "link"), nil, http.StatusBadRequest},
			{"Should reject an invalid cleanup flag", http.MethodDelete, path(clusterQuery+"&cleanupTarget=typo", "link"), nil, http.StatusBadRequest},
			{"Should reject a repeated cleanup flag", http.MethodDelete, path(clusterQuery+"&cleanupTarget=true&cleanupTarget=false", "link"), nil, http.StatusBadRequest},
			{"Should reject a delete without a source cluster", http.MethodDelete, path("cleanupTarget=true", "link"), nil, http.StatusBadRequest},
			{"Should report an unknown cluster", http.MethodGet, path("cluster=" + helpers.GetRandomName("missing")), nil, http.StatusNotFound},
			{"Should reject listing a cluster without a tunnel", http.MethodGet, path(clusterQuery), nil, http.StatusBadRequest},
			{"Should reject cleanup on a cluster without a tunnel", http.MethodDelete, path(clusterQuery+"&cleanupTarget=true", "link"), nil, http.StatusBadRequest},
		}

		for _, c := range cases {
			statusCode, err := helpers.QueryManagementAPI(env, c.method, c.url, c.input, nil, adminHeaders)

			var outcomeErr error
			if statusCode != c.expectStatus {
				outcomeErr = fmt.Errorf("expected status %d, got %d (err: %v)", c.expectStatus, statusCode, err)
			}

			helpers.LogTestOutcome(t, c.condition, outcomeErr)
		}
	}
}

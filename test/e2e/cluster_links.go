package main

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/canonical/lxd/shared/api"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/api/models/v1"
	"github.com/canonical/microcloud-cluster-manager/test/helpers"
)

// testClusterLinksAPI verifies the scope validation, authorization and tunnel handling of the management API cluster
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

		// Reuse the shared admin session. A new login would end it, as each user has one active session.
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

		nonAdminCookies, err := helpers.GetCookies(env, "cluster-manager-e2e-lower-permission@example.org", "cluster-manager-e2e-password")
		if err != nil {
			helpers.LogTestOutcome(t, "Should log in as non-admin", err)
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
			nonAdmin     bool
			expectStatus int
		}{
			{"Should deny a non-admin user listing cluster links", http.MethodGet, path("all-clusters=true"), nil, true, http.StatusForbidden},
			{"Should deny a non-admin user creating a cluster link", http.MethodPost, path(clusterQuery), map[string]any{}, true, http.StatusForbidden},
			{"Should deny a non-admin user updating a cluster link", http.MethodPut, path(clusterQuery, "link"), map[string]any{}, true, http.StatusForbidden},
			{"Should deny a non-admin user deleting a cluster link", http.MethodDelete, path(clusterQuery, "link"), nil, true, http.StatusForbidden},
			{"Should reject a list without a scope", http.MethodGet, path(""), nil, false, http.StatusBadRequest},
			{"Should reject a list with an empty cluster", http.MethodGet, path("cluster="), nil, false, http.StatusBadRequest},
			{"Should reject a list with a repeated cluster", http.MethodGet, path(clusterQuery + "&" + clusterQuery), nil, false, http.StatusBadRequest},
			{"Should reject a list with conflicting scopes", http.MethodGet, path(clusterQuery + "&all-clusters=true"), nil, false, http.StatusBadRequest},
			{"Should reject a list with an invalid all-clusters value", http.MethodGet, path("all-clusters=false"), nil, false, http.StatusBadRequest},
			{"Should reject a list with a malformed query", http.MethodGet, path("cluster=%zz"), nil, false, http.StatusBadRequest},
			{"Should reject a create without a source cluster", http.MethodPost, path(""), map[string]any{}, false, http.StatusBadRequest},
			{"Should reject a create for all clusters", http.MethodPost, path("all-clusters=true"), map[string]any{}, false, http.StatusBadRequest},
			{"Should reject an update with a repeated cluster", http.MethodPut, path(clusterQuery+"&"+clusterQuery, "link"), map[string]any{}, false, http.StatusBadRequest},
			{"Should reject a delete without a source cluster", http.MethodDelete, path("removeReverseEdge=true", "link"), nil, false, http.StatusBadRequest},
			{"Should report an unknown cluster", http.MethodGet, path("cluster=" + helpers.GetRandomName("missing")), nil, false, http.StatusNotFound},
			{"Should reject listing a cluster without a tunnel", http.MethodGet, path(clusterQuery), nil, false, http.StatusBadRequest},
			{"Should reject deleting both edges of a link on a cluster without a tunnel", http.MethodDelete, path(clusterQuery+"&removeReverseEdge=true", "link"), nil, false, http.StatusBadRequest},
		}

		for _, c := range cases {
			headers := adminHeaders
			if c.nonAdmin {
				headers = helpers.AddCookiesToRequest(nonAdminCookies)
			}

			statusCode, err := helpers.QueryManagementAPI(env, c.method, c.url, c.input, nil, headers)

			var outcomeErr error
			if statusCode != c.expectStatus {
				outcomeErr = fmt.Errorf("expected status %d, got %d (err: %v)", c.expectStatus, statusCode, err)
			}

			helpers.LogTestOutcome(t, c.condition, outcomeErr)
		}

		condition = "Should list all clusters without errors for clusters without a tunnel"
		var list models.ClusterLinkList
		statusCode, err := helpers.QueryManagementAPI(env, http.MethodGet, path("all-clusters=true"), nil, &list, adminHeaders)
		switch {
		case err != nil:
			err = fmt.Errorf("status %d: %w", statusCode, err)
		case list.ClusterLinks == nil || list.Errors == nil:
			err = fmt.Errorf("expected non-null rows and errors, got %+v", list)
		default:
			for _, listErr := range list.Errors {
				if listErr.Cluster == clusterName {
					err = fmt.Errorf("cluster without a tunnel was reported as an error: %s", listErr.Error)
				}
			}

			for _, entry := range list.ClusterLinks {
				if entry.SourceCluster == clusterName {
					err = fmt.Errorf("cluster without a tunnel returned link %q", entry.Link.Name)
				}
			}
		}

		helpers.LogTestOutcome(t, condition, err)

		condition = "Should no longer serve the nested cluster links route"
		nested := api.NewURL().Scheme("https").Host(env.ManagementAPIHostPort()).Path("1.0", "remote-cluster", clusterName, "cluster-links")
		statusCode, _ = helpers.QueryManagementAPI(env, http.MethodGet, nested, nil, nil, adminHeaders)
		err = nil
		if statusCode != http.StatusNotFound && statusCode != http.StatusMethodNotAllowed {
			err = fmt.Errorf("expected 404 or 405, got %d", statusCode)
		}

		helpers.LogTestOutcome(t, condition, err)
	}
}

package main

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/canonical/lxd/shared/api"
	"github.com/canonical/microcloud-cluster-manager/test/helpers"
)

// testRemoteClusterLinkPrecheck verifies that the cluster connector validates cluster link
// creation requests and checks both clusters' reachability and permissions before creating
// anything. The fake clusters used in e2e tests never open a tunnel, so a valid request must
// fail fast with an "unreachable" error instead of attempting to create a pending link.
func testRemoteClusterLinkPrecheck(env *helpers.Environment) (testName string, testFunc func(t *testing.T)) {
	return "cluster link creation precheck fails fast without creating anything", func(t *testing.T) {
		sourceName := helpers.GetRandomName("cluster_link_e2e_src")
		targetName := helpers.GetRandomName("cluster_link_e2e_tgt")

		var condition string

		{
			condition = "Should register source and target remote clusters"
			_, err := helpers.RegisterRemoteCluster(env, sourceName)
			if err != nil {
				helpers.LogTestOutcome(t, condition, fmt.Errorf("failed to register source cluster: %w", err))
				return
			}

			_, err = helpers.RegisterRemoteCluster(env, targetName)
			if err != nil {
				helpers.LogTestOutcome(t, condition, fmt.Errorf("failed to register target cluster: %w", err))
				return
			}

			helpers.LogTestOutcome(t, condition, nil)
		}

		defer func() {
			env.RemoveRemoteCluster(sourceName)
			env.RemoveRemoteCluster(targetName)
			env.RemoveRemoteClusterToken(sourceName)
			env.RemoveRemoteClusterToken(targetName)
		}()

		// The internal cluster connector endpoints expect the headers the management API
		// sets when forwarding a user request.
		internalHeaders := func(req *http.Request) error {
			req.Header.Set("Authorization", "Bearer e2e-test-token")
			req.Header.Set("X-User-Secret", "e2e-test-secret")
			return nil
		}

		clusterLinksPath := func(source string) *api.URL {
			return api.NewURL().Scheme("https").Host(env.ClusterConnectorHostPort()).Path("1.0", "remote-cluster", source, "cluster-links")
		}

		type testCase struct {
			condition    string
			source       string
			input        map[string]any
			expectStatus int
			expectInBody string
		}

		cases := []testCase{
			{
				condition:    "Should reject an unsupported cluster link type",
				source:       sourceName,
				input:        map[string]any{"target_cluster": targetName, "type": "bogus"},
				expectStatus: http.StatusBadRequest,
				expectInBody: "Unsupported cluster link type",
			},
			{
				condition:    "Should reject a missing cluster link type",
				source:       sourceName,
				input:        map[string]any{"target_cluster": targetName},
				expectStatus: http.StatusBadRequest,
				expectInBody: "Unsupported cluster link type",
			},
			{
				condition:    "Should reject a target cluster equal to the source cluster",
				source:       sourceName,
				input:        map[string]any{"target_cluster": sourceName, "type": "bidirectional"},
				expectStatus: http.StatusBadRequest,
			},
			{
				condition:    "Should reject a missing target cluster",
				source:       sourceName,
				input:        map[string]any{"type": "bidirectional"},
				expectStatus: http.StatusBadRequest,
			},
			{
				condition:    "Should fail fast when a cluster is unreachable instead of creating a pending link",
				source:       sourceName,
				input:        map[string]any{"target_cluster": targetName, "type": "bidirectional"},
				expectStatus: http.StatusBadGateway,
				expectInBody: "unreachable",
			},
		}

		for _, c := range cases {
			statusCode, err := helpers.QueryClusterConnectorInternal(env, http.MethodPost, clusterLinksPath(c.source), c.input, nil, internalHeaders)

			var outcomeErr error
			if statusCode != c.expectStatus {
				outcomeErr = fmt.Errorf("expected status %d, got %d (err: %v)", c.expectStatus, statusCode, err)
			} else if c.expectInBody != "" && (err == nil || !strings.Contains(err.Error(), c.expectInBody)) {
				outcomeErr = fmt.Errorf("expected error containing %q, got %v", c.expectInBody, err)
			}

			helpers.LogTestOutcome(t, c.condition, outcomeErr)
		}
	}
}

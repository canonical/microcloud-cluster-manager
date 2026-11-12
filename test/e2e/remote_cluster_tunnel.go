package main

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/canonical/lxd/shared/api"
	"github.com/canonical/microcloud-cluster-manager/test/helpers"
)

// tunnelClusterName is a real MicroCloud that is enrolled with the reverse tunnel enabled:
// scripts/mcm-cluster-onboard.sh --count 2 --instance-prefix mcm-e2e.
const tunnelClusterName = "mcm-e2e-01"

// testRemoteClusterTunnel verifies that requests for a real cluster are answered by the
// cluster itself through its tunnel. There is no fallback to a direct connection, so a
// successful cluster link listing proves that the tunnel works.
func testRemoteClusterTunnel(env *helpers.Environment) (testName string, testFunc func(t *testing.T)) {
	return "lxd remote cluster requests are served through the tunnel", func(t *testing.T) {
		var condition string

		{
			condition = "Should register the tunnel"
			deadline := time.Now().Add(3 * time.Minute)
			for {
				cluster, err := helpers.FindRemoteCluster(env, tunnelClusterName)
				if err == nil && cluster.TunnelRegistered {
					break
				}

				if time.Now().After(deadline) {
					if err == nil {
						err = fmt.Errorf("tunnel not registered after 3 minutes")
					}

					helpers.LogTestOutcome(t, condition, err)
				}

				time.Sleep(5 * time.Second)
			}

			helpers.LogTestOutcome(t, condition, nil)
		}

		{
			condition = "Should list the cluster links through the tunnel"
			certPublicKey, err := env.ManagementAPICert().PublicKeyX509()
			if err != nil {
				helpers.LogTestOutcome(t, condition, err)
			}

			headers, err := env.ManagementAPILoginHeaders(certPublicKey)
			if err != nil {
				helpers.LogTestOutcome(t, condition, err)
			}

			var clusterLinks []api.ClusterLink
			path := api.NewURL().Scheme("https").Host(env.ManagementAPIHostPort()).Path("1.0", "remote-cluster", tunnelClusterName, "cluster-links")
			statusCode, err := helpers.QueryManagementAPI(env, http.MethodGet, path, nil, &clusterLinks, headers)
			if err == nil && statusCode != http.StatusOK {
				err = fmt.Errorf("expected status %d, got %d", http.StatusOK, statusCode)
			}

			helpers.LogTestOutcome(t, condition, err)
		}
	}
}

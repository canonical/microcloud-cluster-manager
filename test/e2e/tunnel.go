package main

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/canonical/lxd/shared/api"
	"github.com/canonical/microcloud-cluster-manager/test/helpers"
)

// These real MicroCloud clusters are enrolled with reverse tunnels enabled:
// scripts/mcm-cluster-onboard.sh --count 2 --instance-prefix mcm-e2e.
const (
	tunnelSourceClusterName = "mcm-e2e-01"
	tunnelTargetClusterName = "mcm-e2e-02"
)

// testClusterTunnel verifies that requests for a real cluster are answered by the
// cluster itself through its tunnel. There is no fallback to a direct connection, so a
// successful cluster link listing proves that the tunnel works.
func testClusterTunnel(env *helpers.Environment) (testName string, testFunc func(t *testing.T)) {
	return "lxd remote cluster requests are served through the tunnel", func(t *testing.T) {
		checkClusterTunnel(t, env, tunnelSourceClusterName)
	}
}

// checkClusterTunnel verifies an authenticated MCM cluster-links request, retrying
// only when the cluster has no registered tunnel.
func checkClusterTunnel(t *testing.T, env *helpers.Environment, clusterName string) {
	certPublicKey, err := env.ManagementAPICert().PublicKeyX509()
	helpers.LogTestOutcome(t, "Should read the management API certificate", err)

	headers, err := env.ManagementAPILoginHeaders(certPublicKey)
	helpers.LogTestOutcome(t, "Should log in to the management API", err)

	path := api.NewURL().Scheme("https").Host(env.ManagementAPIHostPort()).Path("1.0", "remote-cluster", clusterName, "cluster-links")
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()

	err = retryUnregisteredTunnel(ctx, clusterName, func() error {
		var links []api.ClusterLink
		status, err := helpers.QueryManagementAPI(env, http.MethodGet, path, nil, &links, headers)
		if err != nil {
			return err
		}
		if status != http.StatusOK || links == nil {
			return fmt.Errorf("unexpected cluster link response (status %d)", status)
		}
		return nil
	})
	helpers.LogTestOutcome(t, fmt.Sprintf("Cluster %q should serve requests through its tunnel", clusterName), err)
}

// retryUnregisteredTunnel retries the request function until the cluster registers its tunnel or the context times out.
// It only retries when the error indicates an unregistered tunnel.
func retryUnregisteredTunnel(ctx context.Context, clusterName string, request func() error) error {
	unregistered := fmt.Sprintf("remote cluster %q does not have an active tunnel", clusterName)
	for {
		err := ctx.Err()
		if err != nil {
			return err
		}

		err = request()
		if err == nil || !api.StatusErrorCheck(err, http.StatusBadRequest) || err.Error() != unregistered {
			return err
		}

		select {
		case <-ctx.Done():
			return fmt.Errorf("cluster %q did not register its tunnel: %w", clusterName, ctx.Err())
		case <-time.After(10 * time.Second):
		}
	}
}

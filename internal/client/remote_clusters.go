// Package client provides a full Go API client.
package client

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/canonical/lxd/shared/api"
	"github.com/canonical/microcluster/client"

	"github.com/canonical/lxd-site-manager/internal/api/types"
)

// GetRemoteClusters gets the remote clusters.
func GetRemoteClusters(ctx context.Context, c *client.Client) ([]types.RemoteCluster, error) {
	queryCtx, cancel := context.WithTimeout(ctx, time.Second*30)
	defer cancel()

	var remoteClusters []types.RemoteCluster
	err := c.Query(queryCtx, http.MethodGet, types.APIVersionPrefix, api.NewURL().Path("remote-clusters"), nil, &remoteClusters)
	if err != nil {
		clientURL := c.URL()
		return nil, fmt.Errorf("Failed performing action on %q: %w", clientURL.String(), err)
	}

	return remoteClusters, nil
}

// GetRemoteCluster gets a remote cluster by name.
func GetRemoteCluster(ctx context.Context, c *client.Client, remoteClusterName string) (*types.RemoteCluster, error) {
	queryCtx, cancel := context.WithTimeout(ctx, time.Second*30)
	defer cancel()

	var remoteCluster types.RemoteCluster
	err := c.Query(queryCtx, http.MethodGet, types.APIVersionPrefix, api.NewURL().Path("remote-clusters", remoteClusterName), nil, &remoteCluster)
	if err != nil {
		clientURL := c.URL()
		return nil, fmt.Errorf("Failed performing action on %q: %w", clientURL.String(), err)
	}

	return &remoteCluster, nil
}

// RemoteClusterPatchCmd sends a client requets to PATCH /1.0/remote-clusters/{remoteClusterName} endpoint.
func RemoteClusterPatchCmd(ctx context.Context, c *client.Client, remoteClusterName string, payload *types.RemoteClusterPatch) error {
	queryCtx, cancel := context.WithTimeout(ctx, time.Second*30)
	defer cancel()

	url := api.NewURL().Path("remote-clusters", remoteClusterName)
	err := c.Query(queryCtx, http.MethodPatch, types.APIVersionPrefix, url, payload, nil)
	if err != nil {
		clientURL := c.URL()
		return fmt.Errorf("Failed performing action on %q: %w", clientURL.String(), err)
	}

	return nil
}

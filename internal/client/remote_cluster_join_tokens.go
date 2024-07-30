package client

import (
	"context"
	"fmt"
	"time"

	"github.com/canonical/lxd/shared/api"
	"github.com/canonical/microcluster/client"

	"github.com/canonical/lxd-cluster-manager/internal/api/types"
)

// RemoteClusterJoinTokenPostCmd sends a POST request to /1.0/remote-cluster-join-token.
func RemoteClusterJoinTokenPostCmd(ctx context.Context, c *client.Client, payload *types.RemoteClusterTokenPost) (token string, err error) {
	queryCtx, cancel := context.WithTimeout(ctx, time.Second*30)
	defer cancel()

	tokenResponse := &types.RemoteClusterTokenPostResponse{}
	url := api.NewURL().Path("remote-cluster-join-token")
	err = c.Query(queryCtx, "POST", types.APIVersionPrefix, url, payload, tokenResponse)
	if err != nil {
		clientURL := c.URL()
		return "", fmt.Errorf("Failed performing action on %q: %w", clientURL.String(), err)
	}

	return tokenResponse.Token, nil
}

// RemoteClusterJoinTokenGetCmd sends a GET request to /1.0/remote-cluster-join-token.
func RemoteClusterJoinTokenGetCmd(ctx context.Context, c *client.Client) ([]types.RemoteClusterToken, error) {
	queryCtx, cancel := context.WithTimeout(ctx, time.Second*30)
	defer cancel()

	tokens := []types.RemoteClusterToken{}
	url := api.NewURL().Path("remote-cluster-join-token")
	err := c.Query(queryCtx, "GET", types.APIVersionPrefix, url, nil, &tokens)
	if err != nil {
		clientURL := c.URL()
		return nil, fmt.Errorf("Failed performing action on %q: %w", clientURL.String(), err)
	}

	return tokens, nil
}

// RemoteClusterJoinTokenDeleteCmd sends a DELETE request to /1.0/remote-cluster-join-token/{remoteClusterName}.
func RemoteClusterJoinTokenDeleteCmd(ctx context.Context, c *client.Client, remoteClusterName string) error {
	queryCtx, cancel := context.WithTimeout(ctx, time.Second*30)
	defer cancel()

	url := api.NewURL().Path("remote-cluster-join-token", remoteClusterName)
	err := c.Query(queryCtx, "DELETE", types.APIVersionPrefix, url, nil, nil)
	if err != nil {
		clientURL := c.URL()
		return fmt.Errorf("Failed performing action on %q: %w", clientURL.String(), err)
	}

	return nil
}

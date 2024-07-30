package client

import (
	"context"
	"fmt"
	"time"

	"github.com/canonical/lxd/shared/api"
	"github.com/canonical/microcluster/client"
	apiTypes "github.com/canonical/microcluster/rest/types"

	"github.com/canonical/lxd-cluster-manager/internal/api/types"
)

// GetDaemonServerConfigs sends a GET request to /daemon/servers for fetching server configs.
func GetDaemonServerConfigs(ctx context.Context, c *client.Client) (map[string]apiTypes.ServerConfig, error) {
	queryCtx, cancel := context.WithTimeout(ctx, time.Second*30)
	defer cancel()

	var serverConfigs map[string]apiTypes.ServerConfig
	endpoint := api.NewURL().Path("daemon", "servers")
	err := c.Query(queryCtx, "GET", types.InternalEndpoint, endpoint, nil, &serverConfigs)
	if err != nil {
		clientURL := c.URL()
		return nil, fmt.Errorf("Failed performing action on %q: %w", clientURL.String(), err)
	}

	return serverConfigs, nil
}

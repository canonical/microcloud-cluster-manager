package client

import (
	"context"
	"fmt"
	"time"

	"github.com/canonical/lxd/shared/api"
	"github.com/canonical/microcluster/client"

	"github.com/canonical/lxd-site-manager/internal/api/types"
)

// ConfigsPatchCmd notifies other members of the cluster that the manager configs have been updated.
func ConfigsPatchCmd(ctx context.Context, c *client.Client, configs *types.ManagerConfigs) error {
	queryCtx, cancel := context.WithTimeout(ctx, time.Second*30)
	defer cancel()

	err := c.Query(queryCtx, "PATCH", types.APIVersionPrefix, api.NewURL().Path("config"), configs, nil)
	if err != nil {
		clientURL := c.URL()
		return fmt.Errorf("Failed performing action on %q: %w", clientURL.String(), err)
	}

	return nil
}

package client

import (
	"context"
	"fmt"
	"time"

	"github.com/canonical/lxd/shared/api"
	"github.com/canonical/microcluster/client"

	"github.com/canonical/lxd-site-manager/internal/api/types"
)

// ManagerConfigsPatchCmd sends a PATCH request to /1.0/config.
func ManagerConfigsPatchCmd(ctx context.Context, c *client.Client, configs *types.ManagerConfigs) error {
	queryCtx, cancel := context.WithTimeout(ctx, time.Second*30)
	defer cancel()

	err := c.Query(queryCtx, "PATCH", types.APIVersionPrefix, api.NewURL().Path("config"), configs, nil)
	if err != nil {
		clientURL := c.URL()
		return fmt.Errorf("Failed performing action on %q: %w", clientURL.String(), err)
	}

	return nil
}

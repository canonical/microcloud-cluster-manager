package client

import (
	"context"
	"fmt"
	"time"

	"github.com/canonical/lxd/shared/api"
	"github.com/canonical/microcluster/client"

	"github.com/canonical/lxd-cluster-manager/internal/api/types"
)

// MemberConfigPatchCmd sends a PATCH request to /1.0/member/{name}/config.
func MemberConfigPatchCmd(ctx context.Context, c *client.Client, member string, configs *types.MemberConfigPatch) error {
	queryCtx, cancel := context.WithTimeout(ctx, time.Second*30)
	defer cancel()

	url := api.NewURL().Path("member", member, "config")
	c.SetClusterNotification()

	err := c.Query(queryCtx, "PATCH", types.APIVersionPrefix, url, configs, nil)
	if err != nil {
		clientURL := c.URL()
		return fmt.Errorf("Failed performing action on %q: %w", clientURL.String(), err)
	}

	return nil
}

// MemberConfigGetCmd sends a GET request to /1.0/member/{name}/config.
func MemberConfigGetCmd(ctx context.Context, c *client.Client, member string) (types.MemberConfig, error) {
	queryCtx, cancel := context.WithTimeout(ctx, time.Second*30)
	defer cancel()

	var memberConfig types.MemberConfig
	url := api.NewURL().Path("member", member, "config")
	err := c.Query(queryCtx, "GET", types.APIVersionPrefix, url, nil, &memberConfig)
	if err != nil {
		clientURL := c.URL()
		return types.MemberConfig{}, fmt.Errorf("Failed performing action on %q: %w", clientURL.String(), err)
	}

	return memberConfig, nil
}

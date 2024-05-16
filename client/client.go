// Package client provides a full Go API client.
package client

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/canonical/lxd/shared/api"
	"github.com/canonical/microcluster/client"

	"github.com/canonical/lxd-site-manager/api/types"
)

// GetSites gets the sites.
func GetSites(ctx context.Context, c *client.Client) ([]types.Site, error) {
	queryCtx, cancel := context.WithTimeout(ctx, time.Second*30)
	defer cancel()

	var sites []types.Site
	err := c.Query(queryCtx, http.MethodGet, api.NewURL().Path("sites"), nil, &sites)
	if err != nil {
		clientURL := c.URL()
		return nil, fmt.Errorf("Failed performing action on %q: %w", clientURL.String(), err)
	}

	return sites, nil
}

// GetSite gets a site by name.
func GetSite(ctx context.Context, c *client.Client, siteName string) (*types.Site, error) {
	queryCtx, cancel := context.WithTimeout(ctx, time.Second*30)
	defer cancel()

	var site types.Site
	err := c.Query(queryCtx, http.MethodGet, api.NewURL().Path("sites", siteName), nil, &site)
	if err != nil {
		clientURL := c.URL()
		return nil, fmt.Errorf("Failed performing action on %q: %w", clientURL.String(), err)
	}

	return &site, nil
}

package client

import (
	"context"
	"fmt"
	"time"

	"github.com/canonical/lxd/shared/api"
	"github.com/canonical/microcluster/client"

	"github.com/canonical/lxd-site-manager/internal/api/types"
)

// ExternalSiteJoinTokenPostCmd sends a POST request to /1.0/external-site-join-token.
func ExternalSiteJoinTokenPostCmd(ctx context.Context, c *client.Client, payload *types.ExternalSiteTokenPost) (token string, err error) {
	queryCtx, cancel := context.WithTimeout(ctx, time.Second*30)
	defer cancel()

	tokenResponse := &types.ExternalSiteTokenPostResponse{}
	url := api.NewURL().Path("external-site-join-token")
	err = c.Query(queryCtx, "POST", types.APIVersionPrefix, url, payload, tokenResponse)
	if err != nil {
		clientURL := c.URL()
		return "", fmt.Errorf("Failed performing action on %q: %w", clientURL.String(), err)
	}

	return tokenResponse.Token, nil
}

// ExternalSiteJoinTokenGetCmd sends a GET request to /1.0/external-site-join-token.
func ExternalSiteJoinTokenGetCmd(ctx context.Context, c *client.Client) ([]types.ExternalSiteToken, error) {
	queryCtx, cancel := context.WithTimeout(ctx, time.Second*30)
	defer cancel()

	tokens := []types.ExternalSiteToken{}
	url := api.NewURL().Path("external-site-join-token")
	err := c.Query(queryCtx, "GET", types.APIVersionPrefix, url, nil, &tokens)
	if err != nil {
		clientURL := c.URL()
		return nil, fmt.Errorf("Failed performing action on %q: %w", clientURL.String(), err)
	}

	return tokens, nil
}

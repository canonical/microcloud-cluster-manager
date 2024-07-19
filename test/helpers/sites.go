package helpers

import (
	"context"
	"net/http"
	"time"

	"github.com/canonical/lxd/shared/api"

	"github.com/canonical/lxd-site-manager/internal/api/types"
)

// FindSite search for a site by name.
func FindSite(env *Environment, siteName string) (*types.Site, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	unixClient, err := NewUnixHTTPClient(env.ControlSocketURL())
	if err != nil {
		return nil, err
	}

	output := &types.Site{}
	path := api.NewURL().Path("1.0", "sites", siteName)
	err = unixClient.Query(ctx, http.MethodGet, path, nil, output, nil)
	if err != nil {
		return nil, err
	}

	return output, nil
}

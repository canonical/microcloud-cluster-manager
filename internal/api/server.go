package api

import (
	"github.com/canonical/microcluster/rest"

	"github.com/canonical/lxd-site-manager/internal/api/types"
)

// Servers contains all the network listeners for site manager.
var Servers = []rest.Server{
	// site management listener (same network listener as core)
	{
		CoreAPI: true,
		Resources: []rest.Resources{
			{
				PathPrefix: types.NoPrefix,
				Endpoints: []rest.Endpoint{
					uiRootCmd,
					uiCmd,
					uiAssetsCmd,
					uiImgCmd,
				},
			},
			{
				PathPrefix: types.ApiVersionPrefix,
				Endpoints: []rest.Endpoint{
					siteCmd,
					sitesCmd,
				},
			},
		},
	},
}

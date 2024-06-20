package api

import (
	"github.com/canonical/microcluster/rest"

	"github.com/canonical/lxd-site-manager/internal/api/types"
)

var siteManagementListener = rest.Server{
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
			PathPrefix: types.APIVersionPrefix,
			Endpoints: []rest.Endpoint{
				siteCmd,
				sitesCmd,
			},
		},
	},
}

// Servers contains all the network listeners for site manager.
var Servers = []rest.Server{
	siteManagementListener,
}

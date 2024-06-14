package api

import (
	"github.com/canonical/microcluster/rest"
)

// Servers contains all the network listeners for site manager.
var Servers = []rest.Server{
	// site management listener (same network listener as core)
	{
		CoreAPI: true,
		Resources: []rest.Resources{
			{
				Path: "",
				Endpoints: []rest.Endpoint{
					uiRootCmd,
					uiCmd,
					uiAssetsCmd,
					uiImgCmd,
				},
			},
			{
				Path: "1.0",
				Endpoints: []rest.Endpoint{
					siteCmd,
					sitesCmd,
				},
			},
		},
	},
}

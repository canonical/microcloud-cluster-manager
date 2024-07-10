package api

import (
	"github.com/canonical/microcluster/rest"

	"github.com/canonical/lxd-site-manager/internal/api/types"
)

var siteManagementListener = rest.Server{
	CoreAPI:   true,
	ServeUnix: true,
	Resources: []rest.Resources{
		{
			PathPrefix: types.NoPrefix,
			Endpoints: append(
				[]rest.Endpoint{
					uiRootCmd,
				},
				generateUIEndpoints()...,
			),
		},
		{
			PathPrefix: types.APIVersionPrefix,
			Endpoints: []rest.Endpoint{
				siteCmd,
				sitesCmd,
				managerConfigsCmd,
				memberConfigCmd,
				memberConfigsCmd,
				externalSiteJoinTokensCmd,
				externalSiteJoinTokenCmd,
			},
		},
	},
}

// GetServers returns all the network listeners for site manager.
func GetServers() map[string]rest.Server {
	return map[string]rest.Server{
		"management-listener": siteManagementListener,
	}
}

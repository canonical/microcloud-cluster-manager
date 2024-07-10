package api

import (
	"github.com/canonical/microcluster/rest"

	"github.com/canonical/lxd-site-manager/internal/api/types"
)

// ListenerName represents the name of any network listener relevant in site manager.
type ListenerName string

const (
	// ManagementListener is the name for the listener for the management API.
	ManagementListener ListenerName = "management-listener"
	// ControlListener is the name for the listener for the control API.
	ControlListener ListenerName = "control-listener"
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

var siteControlListener = rest.Server{
	CoreAPI:   false,
	PreInit:   false,
	ServeUnix: false,
	Resources: []rest.Resources{
		{
			PathPrefix: types.APIVersionPrefix,
			Endpoints: []rest.Endpoint{
				sitesControlCmd,
			},
		},
	},
}

// GetServers returns all the network listeners for site manager.
func GetServers() map[string]rest.Server {
	return map[string]rest.Server{
		string(ManagementListener): siteManagementListener,
		string(ControlListener):    siteControlListener,
	}
}

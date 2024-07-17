package api

import (
	"github.com/canonical/microcluster/rest"

	"github.com/canonical/lxd-site-manager/internal/api/types"
	"github.com/canonical/lxd-site-manager/internal/state"
)

// ListenerName represents the name of any network listener relevant in site manager.
type ListenerName string

const (
	// ManagementListener is the name for the listener for the management API.
	ManagementListener ListenerName = "management-listener"
	// ControlListener is the name for the listener for the control API.
	ControlListener ListenerName = "control-listener"
)

func siteManagementListener(s *state.SiteManagerState) rest.Server {
	return rest.Server{
		CoreAPI:   true,
		ServeUnix: true,
		Resources: []rest.Resources{
			{
				PathPrefix: types.NoPrefix,
				Endpoints: append(
					[]rest.Endpoint{
						uiRootCmd,
						oidcLoginCmd(s),
						oidcCallbackCmd(s),
						oidcLogoutCmd(s),
					},
					generateUIEndpoints()...,
				),
			},
			{
				PathPrefix: types.APIVersionPrefix,
				Endpoints: []rest.Endpoint{
					apiRootCmd(s),
					siteCmd(s),
					sitesCmd(s),
					managerConfigsCmd(s),
					memberConfigCmd(s),
					memberConfigsCmd(s),
					externalSiteJoinTokensCmd(s),
					externalSiteJoinTokenCmd(s),
				},
			},
		},
	}
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
				sitesStatusCmd,
			},
		},
	},
}

// GetServers returns all the network listeners for site manager.
func GetServers(s *state.SiteManagerState) map[string]rest.Server {
	return map[string]rest.Server{
		string(ManagementListener): siteManagementListener(s),
		string(ControlListener):    siteControlListener,
	}
}

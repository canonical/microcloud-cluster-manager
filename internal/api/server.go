package api

import (
	"github.com/canonical/microcluster/rest"

	"github.com/canonical/lxd-cluster-manager/internal/api/types"
	"github.com/canonical/lxd-cluster-manager/internal/state"
)

// ListenerName represents the name of any network listener relevant in Cluster Manager.
type ListenerName string

const (
	// ManagementListener is the name for the listener for the management API.
	ManagementListener ListenerName = "management-listener"
	// ControlListener is the name for the listener for the control API.
	ControlListener ListenerName = "control-listener"
)

func remoteClusterManagementListener(s *state.ClusterManagerState) rest.Server {
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
					remoteClusterCmd(s),
					remoteClustersCmd(s),
					managerConfigsCmd(s),
					memberConfigCmd(s),
					memberConfigsCmd(s),
					remoteClusterJoinTokensCmd(s),
					remoteClusterJoinTokenCmd(s),
				},
			},
		},
	}
}

var remoteClusterControlListener = rest.Server{
	CoreAPI:   false,
	PreInit:   false,
	ServeUnix: false,
	Resources: []rest.Resources{
		{
			PathPrefix: types.APIVersionPrefix,
			Endpoints: []rest.Endpoint{
				remoteClustersControlCmd,
				remoteClustersStatusCmd,
			},
		},
	},
}

// GetServers returns all the network listeners for Cluster Manager.
func GetServers(s *state.ClusterManagerState) map[string]rest.Server {
	return map[string]rest.Server{
		string(ManagementListener): remoteClusterManagementListener(s),
		string(ControlListener):    remoteClusterControlListener,
	}
}

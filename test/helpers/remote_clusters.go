package helpers

import (
	"context"
	"net/http"
	"time"

	"github.com/canonical/lxd/shared/api"

	"github.com/canonical/lxd-site-manager/internal/api/types"
)

// FindRemoteCluster search for a remote cluster by name.
func FindRemoteCluster(env *Environment, remoteClusterName string) (*types.RemoteCluster, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	unixClient, err := NewUnixHTTPClient(env.ControlSocketURL())
	if err != nil {
		return nil, err
	}

	output := &types.RemoteCluster{}
	path := api.NewURL().Path("1.0", "remote-clusters", remoteClusterName)
	err = unixClient.Query(ctx, http.MethodGet, path, nil, output, nil)
	if err != nil {
		return nil, err
	}

	return output, nil
}

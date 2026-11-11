package v1

import (
	"net/http"
	"net/url"

	"github.com/canonical/lxd/lxd/response"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/request"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/types"
)

func remoteClusterReplicatorsGet(rc types.RouteConfig) types.EndpointHandler {
	return forwardReplicatorToClusterConnector(rc, "")
}

func remoteClusterReplicatorsPost(rc types.RouteConfig) types.EndpointHandler {
	return forwardReplicatorToClusterConnector(rc, "")
}

func remoteClusterReplicatorGet(rc types.RouteConfig) types.EndpointHandler {
	return forwardReplicatorToClusterConnector(rc, "")
}

func remoteClusterReplicatorPut(rc types.RouteConfig) types.EndpointHandler {
	return forwardReplicatorToClusterConnector(rc, "")
}

func remoteClusterReplicatorDelete(rc types.RouteConfig) types.EndpointHandler {
	return forwardReplicatorToClusterConnector(rc, "")
}

func remoteClusterReplicatorStateGet(rc types.RouteConfig) types.EndpointHandler {
	return forwardReplicatorToClusterConnector(rc, "/state")
}

func remoteClusterReplicatorStatePut(rc types.RouteConfig) types.EndpointHandler {
	return forwardReplicatorToClusterConnector(rc, "/state")
}

// forwardReplicatorToClusterConnector forwards a replicator request to the cluster connector holding the tunnel.
// Only a given project query parameter is forwarded, the cluster connector validates it.
func forwardReplicatorToClusterConnector(rc types.RouteConfig, pathSuffix string) types.EndpointHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		clusterConnectorPath := "/replicators"

		replicatorName, hasReplicatorName, err := request.OptionalPathParam(r, request.ReplicatorNameParam, "replicator name")
		if err != nil {
			return response.SmartError(err).Render(w, r)
		}

		if hasReplicatorName {
			clusterConnectorPath += "/" + url.PathEscape(replicatorName)
		}

		clusterConnectorPath += pathSuffix

		project := request.QueryParam(r, "project")
		if project != "" {
			query := url.Values{}
			query.Set("project", project)
			clusterConnectorPath += "?" + query.Encode()
		}

		return forwardToClusterConnector(rc, w, r, clusterConnectorPath)
	}
}

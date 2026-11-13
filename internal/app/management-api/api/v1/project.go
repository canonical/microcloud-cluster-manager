package v1

import (
	"net/http"
	"net/url"

	"github.com/canonical/lxd/lxd/response"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/request"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/types"
)

func remoteClusterProjectsGet(rc types.RouteConfig) types.EndpointHandler {
	return forwardProjectToClusterConnector(rc, "")
}

func remoteClusterProjectsPost(rc types.RouteConfig) types.EndpointHandler {
	return forwardProjectToClusterConnector(rc, "")
}

func remoteClusterProjectGet(rc types.RouteConfig) types.EndpointHandler {
	return forwardProjectToClusterConnector(rc, "")
}

func remoteClusterProjectPut(rc types.RouteConfig) types.EndpointHandler {
	return forwardProjectToClusterConnector(rc, "")
}

func remoteClusterProjectStateGet(rc types.RouteConfig) types.EndpointHandler {
	return forwardProjectToClusterConnector(rc, "/state")
}

func remoteClusterProjectStatePut(rc types.RouteConfig) types.EndpointHandler {
	return forwardProjectToClusterConnector(rc, "/state")
}

// forwardProjectToClusterConnector forwards a project request to the cluster connector holding the tunnel.
// Only a given force query parameter is forwarded, the cluster connector validates it.
func forwardProjectToClusterConnector(rc types.RouteConfig, pathSuffix string) types.EndpointHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		clusterConnectorPath := "/projects"

		projectName, hasProjectName, err := request.OptionalPathParam(r, request.ProjectNameParam, "project name")
		if err != nil {
			return response.SmartError(err).Render(w, r)
		}

		if hasProjectName {
			clusterConnectorPath += "/" + url.PathEscape(projectName)
		}

		clusterConnectorPath += pathSuffix

		force := request.QueryParam(r, "force")
		if force != "" {
			query := url.Values{}
			query.Set("force", force)
			clusterConnectorPath += "?" + query.Encode()
		}

		return forwardToClusterConnector(rc, w, r, clusterConnectorPath)
	}
}

package v1

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/canonical/lxd/lxd/response"
	"github.com/canonical/lxd/shared/api"
	"github.com/canonical/microcloud-cluster-manager/internal/app/cluster-connector/core/project"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/request"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/types"
)

func remoteClusterProjectsGet(rc types.RouteConfig) types.EndpointHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		remoteClusterName, err := request.RemoteClusterName(r)
		if err != nil {
			return response.SmartError(err).Render(w, r)
		}

		return forwardTunnelRequest(rc, w, r, http.MethodGet, project.ListPath(), remoteClusterName, nil, "list projects")
	}
}

func remoteClusterProjectsPost(rc types.RouteConfig) types.EndpointHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		remoteClusterName, err := request.RemoteClusterName(r)
		if err != nil {
			return response.SmartError(err).Render(w, r)
		}

		var projectRequest api.ProjectsPost
		err = decodeBody(r, &projectRequest)
		if err != nil {
			return response.SmartError(err).Render(w, r)
		}

		if projectRequest.Name == "" {
			return response.BadRequest(errors.New("Missing project name")).Render(w, r)
		}

		body, err := json.Marshal(projectRequest)
		if err != nil {
			return response.InternalError(fmt.Errorf("Failed to encode project request: %w", err)).Render(w, r)
		}

		return forwardTunnelRequest(rc, w, r, http.MethodPost, project.CreatePath(), remoteClusterName, body, "create project")
	}
}

func remoteClusterProjectGet(rc types.RouteConfig) types.EndpointHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		remoteClusterName, projectName, err := getRemoteClusterAndProjectName(r)
		if err != nil {
			return response.SmartError(err).Render(w, r)
		}

		return forwardTunnelRequest(rc, w, r, http.MethodGet, project.Path(projectName), remoteClusterName, nil, "get project")
	}
}

func remoteClusterProjectPut(rc types.RouteConfig) types.EndpointHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		remoteClusterName, projectName, err := getRemoteClusterAndProjectName(r)
		if err != nil {
			return response.SmartError(err).Render(w, r)
		}

		var projectRequest api.ProjectPut
		err = decodeBody(r, &projectRequest)
		if err != nil {
			return response.SmartError(err).Render(w, r)
		}

		body, err := json.Marshal(projectRequest)
		if err != nil {
			return response.InternalError(fmt.Errorf("Failed to encode project request: %w", err)).Render(w, r)
		}

		return forwardTunnelRequest(rc, w, r, http.MethodPut, project.Path(projectName), remoteClusterName, body, "update project")
	}
}

func remoteClusterProjectStateGet(rc types.RouteConfig) types.EndpointHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		remoteClusterName, projectName, err := getRemoteClusterAndProjectName(r)
		if err != nil {
			return response.SmartError(err).Render(w, r)
		}

		return forwardTunnelRequest(rc, w, r, http.MethodGet, project.StatePath(projectName, false), remoteClusterName, nil, "get project state")
	}
}

// remoteClusterProjectStatePut updates the replica mode of a project. LXD answers with an asynchronous operation
// on the remote cluster, which is returned as is without waiting for it to complete.
func remoteClusterProjectStatePut(rc types.RouteConfig) types.EndpointHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		remoteClusterName, projectName, err := getRemoteClusterAndProjectName(r)
		if err != nil {
			return response.SmartError(err).Render(w, r)
		}

		force, err := request.BoolQueryParam(r, "force")
		if err != nil {
			return response.SmartError(err).Render(w, r)
		}

		var stateRequest api.ProjectStatePut
		err = decodeBody(r, &stateRequest)
		if err != nil {
			return response.SmartError(err).Render(w, r)
		}

		body, err := json.Marshal(stateRequest)
		if err != nil {
			return response.InternalError(fmt.Errorf("Failed to encode project state request: %w", err)).Render(w, r)
		}

		return forwardTunnelRequest(rc, w, r, http.MethodPut, project.StatePath(projectName, force), remoteClusterName, body, "update project state")
	}
}

// getRemoteClusterAndProjectName reads the remote cluster name and the required project name from the request path.
func getRemoteClusterAndProjectName(r *http.Request) (string, string, error) {
	remoteClusterName, err := request.RemoteClusterName(r)
	if err != nil {
		return "", "", err
	}

	projectName, err := request.RequiredPathParam(r, request.ProjectNameParam, "project name")
	if err != nil {
		return "", "", err
	}

	return remoteClusterName, projectName, nil
}

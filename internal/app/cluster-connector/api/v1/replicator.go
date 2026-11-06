package v1

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/canonical/lxd/lxd/response"
	"github.com/canonical/lxd/shared/api"
	"github.com/canonical/microcloud-cluster-manager/internal/app/cluster-connector/core/replicator"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/types"
	"github.com/gorilla/mux"
)

// replicatorScope identifies the remote cluster and project a replicator request targets.
type replicatorScope struct {
	remoteClusterName string
	project           string
}

func remoteClusterReplicatorsGet(rc types.RouteConfig) types.EndpointHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		scope, err := getReplicatorScope(r, false)
		if err != nil {
			return response.SmartError(err).Render(w, r)
		}

		return forwardReplicatorRequest(rc, w, r, scope, http.MethodGet, replicator.ListPath(scope.project), nil, "list replicators")
	}
}

func remoteClusterReplicatorsPost(rc types.RouteConfig) types.EndpointHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		scope, err := getReplicatorScope(r, true)
		if err != nil {
			return response.SmartError(err).Render(w, r)
		}

		var replicatorRequest api.ReplicatorsPost
		err = decodeBody(r, &replicatorRequest)
		if err != nil {
			return response.SmartError(err).Render(w, r)
		}

		if replicatorRequest.Name == "" {
			return response.BadRequest(errors.New("Missing replicator name")).Render(w, r)
		}

		body, err := json.Marshal(replicatorRequest)
		if err != nil {
			return response.InternalError(fmt.Errorf("Failed to encode replicator request: %w", err)).Render(w, r)
		}

		return forwardReplicatorRequest(rc, w, r, scope, http.MethodPost, replicator.CollectionPath(scope.project), body, "create replicator")
	}
}

func remoteClusterReplicatorGet(rc types.RouteConfig) types.EndpointHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		scope, replicatorName, err := getReplicatorScopeAndName(r)
		if err != nil {
			return response.SmartError(err).Render(w, r)
		}

		return forwardReplicatorRequest(rc, w, r, scope, http.MethodGet, replicator.Path(scope.project, replicatorName), nil, "get replicator")
	}
}

func remoteClusterReplicatorPut(rc types.RouteConfig) types.EndpointHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		scope, replicatorName, err := getReplicatorScopeAndName(r)
		if err != nil {
			return response.SmartError(err).Render(w, r)
		}

		var replicatorRequest api.ReplicatorPut
		err = decodeBody(r, &replicatorRequest)
		if err != nil {
			return response.SmartError(err).Render(w, r)
		}

		body, err := json.Marshal(replicatorRequest)
		if err != nil {
			return response.InternalError(fmt.Errorf("Failed to encode replicator request: %w", err)).Render(w, r)
		}

		return forwardReplicatorRequest(rc, w, r, scope, http.MethodPut, replicator.Path(scope.project, replicatorName), body, "update replicator")
	}
}

func remoteClusterReplicatorDelete(rc types.RouteConfig) types.EndpointHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		scope, replicatorName, err := getReplicatorScopeAndName(r)
		if err != nil {
			return response.SmartError(err).Render(w, r)
		}

		return forwardReplicatorRequest(rc, w, r, scope, http.MethodDelete, replicator.Path(scope.project, replicatorName), nil, "delete replicator")
	}
}

func remoteClusterReplicatorStateGet(rc types.RouteConfig) types.EndpointHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		scope, replicatorName, err := getReplicatorScopeAndName(r)
		if err != nil {
			return response.SmartError(err).Render(w, r)
		}

		return forwardReplicatorRequest(rc, w, r, scope, http.MethodGet, replicator.StatePath(scope.project, replicatorName), nil, "get replicator state")
	}
}

// remoteClusterReplicatorStatePut runs a replicator. LXD answers with an asynchronous operation on the
// remote cluster, which is returned as is without waiting for the run to complete.
func remoteClusterReplicatorStatePut(rc types.RouteConfig) types.EndpointHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		scope, replicatorName, err := getReplicatorScopeAndName(r)
		if err != nil {
			return response.SmartError(err).Render(w, r)
		}

		var stateRequest api.ReplicatorStatePut
		err = decodeBody(r, &stateRequest)
		if err != nil {
			return response.SmartError(err).Render(w, r)
		}

		err = replicator.ValidateAction(stateRequest.Action)
		if err != nil {
			return response.BadRequest(err).Render(w, r)
		}

		body, err := json.Marshal(stateRequest)
		if err != nil {
			return response.InternalError(fmt.Errorf("Failed to encode replicator state request: %w", err)).Render(w, r)
		}

		return forwardReplicatorRequest(rc, w, r, scope, http.MethodPut, replicator.StatePath(scope.project, replicatorName), body, "run replicator")
	}
}

// getReplicatorScope reads the remote cluster from the path and the project from the query.
// An empty project is only allowed if the project is not required.
func getReplicatorScope(r *http.Request, requireProject bool) (replicatorScope, error) {
	remoteClusterName, err := url.PathUnescape(mux.Vars(r)["remoteClusterName"])
	if err != nil {
		return replicatorScope{}, api.StatusErrorf(http.StatusBadRequest, "Invalid remote cluster name: %v", err)
	}

	project := r.URL.Query().Get("project")
	if requireProject && project == "" {
		return replicatorScope{}, api.StatusErrorf(http.StatusBadRequest, "Missing project")
	}

	return replicatorScope{remoteClusterName: remoteClusterName, project: project}, nil
}

// getReplicatorScopeAndName reads the replicator scope and the required replicator name from the request.
func getReplicatorScopeAndName(r *http.Request) (replicatorScope, string, error) {
	scope, err := getReplicatorScope(r, true)
	if err != nil {
		return replicatorScope{}, "", err
	}

	replicatorName, err := url.PathUnescape(mux.Vars(r)["replicatorName"])
	if err != nil {
		return replicatorScope{}, "", api.StatusErrorf(http.StatusBadRequest, "Invalid replicator name: %v", err)
	}

	if replicatorName == "" {
		return replicatorScope{}, "", api.StatusErrorf(http.StatusBadRequest, "Missing replicator name")
	}

	return scope, replicatorName, nil
}

// forwardReplicatorRequest verifies that the remote cluster supports replicators, sends the request over
// its tunnel and writes the LXD response.
func forwardReplicatorRequest(rc types.RouteConfig, w http.ResponseWriter, r *http.Request, scope replicatorScope, method string, path string, body []byte, operation string) error {
	err := checkTunnelRequirements(rc, r, scope.remoteClusterName, replicator.Requirements())
	if err != nil {
		return response.SmartError(err).Render(w, r)
	}

	return forwardTunnelRequest(rc, w, r, method, path, scope.remoteClusterName, body, operation)
}

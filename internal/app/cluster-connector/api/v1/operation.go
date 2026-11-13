package v1

import (
	"net/http"

	"github.com/canonical/lxd/lxd/response"
	"github.com/canonical/microcloud-cluster-manager/internal/app/cluster-connector/core/operation"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/request"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/types"
)

// remoteClusterOperationGet returns an operation of the remote cluster, e.g. to track an asynchronous project state update.
func remoteClusterOperationGet(rc types.RouteConfig) types.EndpointHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		remoteClusterName, err := request.RemoteClusterName(r)
		if err != nil {
			return response.SmartError(err).Render(w, r)
		}

		operationID, err := request.RequiredPathParam(r, request.OperationIDParam, "operation ID")
		if err != nil {
			return response.SmartError(err).Render(w, r)
		}

		return forwardTunnelRequest(rc, w, r, http.MethodGet, operation.Path(operationID), remoteClusterName, nil, "get operation")
	}
}

package v1

import (
	"net/http"
	"net/url"

	"github.com/canonical/lxd/lxd/response"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/request"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/types"
)

func remoteClusterOperationGet(rc types.RouteConfig) types.EndpointHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		operationID, err := request.RequiredPathParam(r, request.OperationIDParam, "operation ID")
		if err != nil {
			return response.SmartError(err).Render(w, r)
		}

		return forwardToClusterConnector(rc, w, r, "/operations/"+url.PathEscape(operationID))
	}
}

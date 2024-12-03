package v1

import (
	"fmt"
	"net/http"

	"github.com/canonical/lxd-cluster-manager/internal/pkg/logger"
	"github.com/canonical/lxd-cluster-manager/internal/pkg/types"
	"github.com/canonical/lxd/lxd/response"
)

// Status is the readiness / liveliness check endpoint.
var Status = types.RouteGroup{
	Prefix: "status",
	Endpoints: []types.Endpoint{
		{
			Method:  http.MethodGet,
			Handler: statusGet,
		},
	},
}

func statusGet(rc types.RouteConfig) types.EndpointHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		// ensure the database is ready
		err := rc.DB.StatusCheck(r.Context())
		if err != nil {
			logger.Log.Errorw("status check failed", "status", "database unreachable", "ERROR", err)
			return response.InternalError(fmt.Errorf("database connection error: %w", err)).Render(w, r)
		}

		return response.SyncResponse(true, nil).Render(w, r)
	}
}

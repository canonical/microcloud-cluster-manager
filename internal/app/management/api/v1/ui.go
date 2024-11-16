package v1

import (
	"net/http"

	"github.com/canonical/lxd-cluster-manager/internal/pkg/database"
	"github.com/canonical/lxd-cluster-manager/internal/pkg/types"
	"github.com/canonical/lxd/lxd/response"
	"go.uber.org/zap"
)

var UI = types.RouteGroup{
	IsRoot: true,
	Prefix: "ui",
	Endpoints: []types.Endpoint{
		{
			Path:    "ui",
			Method:  http.MethodGet,
			Handler: ui,
		},
	},
}

func ui(db *database.DB, logger *zap.SugaredLogger) types.EndpointHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		response.SyncResponse(true, "ui").Render(w, r)
		return nil
	}
}

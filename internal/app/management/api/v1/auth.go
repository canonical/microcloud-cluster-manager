package v1

import (
	"net/http"

	"github.com/canonical/lxd-cluster-manager/internal/pkg/database"
	"github.com/canonical/lxd-cluster-manager/internal/pkg/types"
	"github.com/canonical/lxd/lxd/response"
	"go.uber.org/zap"
)

var Auth = types.RouteGroup{
	IsRoot: true,
	Prefix: "oidc",
	Endpoints: []types.Endpoint{
		{
			Path:    "login",
			Method:  http.MethodGet,
			Handler: login,
		},
		{
			Path:    "callback",
			Method:  http.MethodGet,
			Handler: callback,
		},
		{
			Path:    "logout",
			Method:  http.MethodGet,
			Handler: logout,
		},
	},
}

func login(db *database.DB, logger *zap.SugaredLogger) types.EndpointHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		response.SyncResponse(true, "oidc/login").Render(w, r)
		return nil
	}
}

func callback(db *database.DB, logger *zap.SugaredLogger) types.EndpointHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		response.SyncResponse(true, "oidc/callback").Render(w, r)
		return nil
	}
}

func logout(db *database.DB, logger *zap.SugaredLogger) types.EndpointHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		response.SyncResponse(true, "oidc/logout").Render(w, r)
		return nil
	}
}

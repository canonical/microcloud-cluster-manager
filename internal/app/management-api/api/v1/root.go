package v1

import (
	"fmt"
	"net/http"

	"github.com/canonical/lxd/lxd/response"
	"github.com/canonical/microcloud-cluster-manager/internal/app/management-api/core/auth"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/types"
)

// APIRoot is the API root endpoint.
var APIRoot = types.RouteGroup{
	Prefix: "",
	Middlewares: []types.RouteMiddleware{
		auth.AuthMiddleware,
	},
	Endpoints: []types.Endpoint{
		{
			Method:            http.MethodGet,
			Handler:           apiRootGet,
			AllowUnauthorized: true,
		},
	},
}

func apiRootGet(rc types.RouteConfig) types.EndpointHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		userInfo, ok := r.Context().Value(types.UserInfoKey).(*types.UserInfo)
		if !ok {
			return response.InternalError(fmt.Errorf("failed to get user information from the request context")).Render(w, r)
		}

		systemInfo := types.APIRoot{
			UserInfo: *userInfo,
			Trusted:  true,
		}

		return response.SyncResponse(true, systemInfo).Render(w, r)
	}
}

package api

import (
	"fmt"
	"net/http"

	"github.com/canonical/lxd/lxd/response"
	"github.com/canonical/microcluster/v2/rest"
	microState "github.com/canonical/microcluster/v2/state"

	"github.com/canonical/lxd-cluster-manager/internal/api/types"
	"github.com/canonical/lxd-cluster-manager/internal/state"
)

func apiRootCmd(s *state.ClusterManagerState) rest.Endpoint {
	return rest.Endpoint{
		Path: "",
		Get: rest.EndpointAction{
			Handler:        apiRootGet,
			AllowUntrusted: true,
			AccessHandler:  oidcAuthHandler(s),
		},
	}
}

func apiRootGet(s microState.State, r *http.Request) response.Response {
	userInfo, ok := r.Context().Value(types.UserInfoKey).(*types.UserInfo)
	if !ok {
		return response.InternalError(fmt.Errorf("failed to get user information from the request context"))
	}

	systemInfo := types.APIRoot{
		UserInfo: *userInfo,
		Trusted:  true,
	}

	return response.SyncResponse(true, systemInfo)
}

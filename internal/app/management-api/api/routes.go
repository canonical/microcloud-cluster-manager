package routes

import (
	apiV1 "github.com/canonical/lxd-cluster-manager/internal/app/management-api/api/v1"
	"github.com/canonical/lxd-cluster-manager/internal/pkg/types"
)

// APIRoutes defines the API routes for the management API.
var APIRoutes = []types.RouteGroup{
	apiV1.Status,
	apiV1.UI,
	apiV1.UIRoot,
	apiV1.APIRoot,
	apiV1.Auth,
	apiV1.Configuration,
	apiV1.RemoteCluster,
	apiV1.RemoteClusterJoinToken,
}

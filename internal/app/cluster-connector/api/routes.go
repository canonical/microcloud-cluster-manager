package routes

import (
	apiV1 "github.com/canonical/lxd-cluster-manager/internal/app/cluster-connector/api/v1"
	"github.com/canonical/lxd-cluster-manager/internal/pkg/types"
)

// APIRoutes is the list of all the API routes for the cluster-connector service.
var APIRoutes = []types.RouteGroup{
	apiV1.Status,
	apiV1.RemoteCluster,
	apiV1.RemoteClusterProtected,
}

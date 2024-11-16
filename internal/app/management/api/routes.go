package routes

import (
	v1 "github.com/canonical/lxd-cluster-manager/internal/app/management/api/v1"
	"github.com/canonical/lxd-cluster-manager/internal/pkg/types"
)

var APIRoutes = []types.RouteGroup{
	v1.UI,
	v1.Auth,
	v1.RemoteCluster,
	v1.RemoteClusterJoinToken,
}

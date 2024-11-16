package routes

import (
	v1 "github.com/canonical/lxd-cluster-manager/internal/app/control/api/v1"
	"github.com/canonical/lxd-cluster-manager/internal/pkg/types"
)

var APIRoutes = []types.RouteGroup{
	v1.RemoteCluster,
}

package v1

import (
	"context"
	"net/http"
	"time"

	"github.com/canonical/lxd/lxd/response"
	"github.com/canonical/lxd/shared"
	"github.com/canonical/microcloud-cluster-manager/internal/app/management-api/core/auth"
	"github.com/canonical/microcloud-cluster-manager/internal/app/management-api/core/cluster_link"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/database/store"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/types"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
)

var ClusterLinks = types.RouteGroup{
	Prefix: "cluster-links",
	Middlewares: []types.RouteMiddleware{
		auth.AuthMiddleware,
	},
	Endpoints: []types.Endpoint{
		{
			Method:  http.MethodGet,
			Handler: clusterLinksGet,
		},
		{
			Method:  http.MethodPost,
			Handler: clusterLinksPost,
		},
		{
			Path:    "{clusterLinkName}",
			Method:  http.MethodPut,
			Handler: clusterLinkPut,
		},
		{
			Path:    "{clusterLinkName}",
			Method:  http.MethodDelete,
			Handler: clusterLinkDelete,
		},
	},
}

// clusterLinksConnectorPath is the cluster connector path of the cluster links of a managed cluster.
const clusterLinksConnectorPath = "/cluster-links"

func clusterLinksGet(rc types.RouteConfig) types.EndpointHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		scope, err := cluster_link.ParseListScope(r.URL.RawQuery)
		if err != nil {
			return response.SmartError(err).Render(w, r)
		}

		// Targets are resolved against all managed clusters, also for the links of a single cluster.
		clusters, err := loadRemoteClusters(rc, r)
		if err != nil {
			return response.SmartError(err).Render(w, r)
		}

		if scope.AllClusters {
			options := cluster_link.DefaultListOptions(time.Duration(rc.Env.WriteTimeout) * time.Second)
			list := listAllClusterLinks(rc, r, clusters, options)
			return response.SyncResponse(true, list).Render(w, r)
		}

		list, err := listClusterLinks(rc, r, clusters, scope.Cluster)
		if err != nil {
			return response.SmartError(err).Render(w, r)
		}

		return response.SyncResponse(true, list).Render(w, r)
	}
}

func clusterLinksPost(rc types.RouteConfig) types.EndpointHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		clusterName, err := cluster_link.ParseSourceCluster(r.URL.RawQuery)
		if err != nil {
			return response.SmartError(err).Render(w, r)
		}

		return forwardToRemoteClusterConnector(rc, w, r, clusterName, clusterLinksConnectorPath)
	}
}

func clusterLinkPut(rc types.RouteConfig) types.EndpointHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		clusterName, err := cluster_link.ParseSourceCluster(r.URL.RawQuery)
		if err != nil {
			return response.SmartError(err).Render(w, r)
		}

		return forwardToRemoteClusterConnector(rc, w, r, clusterName, cluster_link.LinkPath(mux.Vars(r)["clusterLinkName"]))
	}
}

func clusterLinkDelete(rc types.RouteConfig) types.EndpointHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		clusterName, err := cluster_link.ParseSourceCluster(r.URL.RawQuery)
		if err != nil {
			return response.SmartError(err).Render(w, r)
		}

		clusterLinkName := mux.Vars(r)["clusterLinkName"]
		removeReverseEdge := shared.IsTrue(r.URL.Query().Get("removeReverseEdge"))
		if !removeReverseEdge {
			return forwardToRemoteClusterConnector(rc, w, r, clusterName, cluster_link.LinkPath(clusterLinkName))
		}

		clusters, err := loadRemoteClusters(rc, r)
		if err != nil {
			return response.SmartError(err).Render(w, r)
		}

		err = deleteClusterLinkOnBothClusters(rc, r, clusters, clusterName, clusterLinkName)
		if err != nil {
			return response.SmartError(err).Render(w, r)
		}

		return response.EmptySyncResponse.Render(w, r)
	}
}

// loadRemoteClusters returns all managed clusters with their tunnel details.
func loadRemoteClusters(rc types.RouteConfig, r *http.Request) ([]store.RemoteClusterWithDetail, error) {
	var clusters []store.RemoteClusterWithDetail
	err := rc.DB.Transaction(r.Context(), func(ctx context.Context, tx *sqlx.Tx) error {
		var err error
		clusters, err = store.GetRemoteClustersWithDetails(ctx, tx)
		return err
	})

	return clusters, err
}

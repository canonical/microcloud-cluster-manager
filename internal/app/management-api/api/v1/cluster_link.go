package v1

import (
	"context"
	"net/http"
	"net/url"
	"time"

	"github.com/canonical/lxd/lxd/response"
	"github.com/canonical/microcloud-cluster-manager/internal/app/management-api/core/auth"
	"github.com/canonical/microcloud-cluster-manager/internal/app/management-api/core/cluster_link"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/database/store"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/request"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/types"
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

// clusterLinksConnectorPath returns the top-level connector path scoped to one cluster.
// An empty link name selects the collection.
func clusterLinksConnectorPath(clusterName string, linkName string) string {
	path := "/1.0/cluster-links"
	if linkName != "" {
		path += "/" + url.PathEscape(linkName)
	}
	return path + "?cluster=" + url.QueryEscape(clusterName)
}

func clusterLinksGet(rc types.RouteConfig) types.EndpointHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		scope, err := request.ParseListScope(r.URL.RawQuery)
		if err != nil {
			return response.SmartError(err).Render(w, r)
		}

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
		clusterName, err := request.ParseSourceCluster(r.URL.RawQuery)
		if err != nil {
			return response.SmartError(err).Render(w, r)
		}

		return forwardToRemoteClusterConnector(rc, w, r, clusterName, clusterLinksConnectorPath(clusterName, ""))
	}
}

func clusterLinkPut(rc types.RouteConfig) types.EndpointHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		clusterName, err := request.ParseSourceCluster(r.URL.RawQuery)
		if err != nil {
			return response.SmartError(err).Render(w, r)
		}

		linkName, err := request.RequiredPathParam(r, request.ClusterLinkNameParam, "cluster link name")
		if err != nil {
			return response.SmartError(err).Render(w, r)
		}

		return forwardToRemoteClusterConnector(rc, w, r, clusterName, clusterLinksConnectorPath(clusterName, linkName))
	}
}

func clusterLinkDelete(rc types.RouteConfig) types.EndpointHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		clusterName, err := request.ParseSourceCluster(r.URL.RawQuery)
		if err != nil {
			return response.SmartError(err).Render(w, r)
		}

		clusterLinkName, err := request.RequiredPathParam(r, request.ClusterLinkNameParam, "cluster link name")
		if err != nil {
			return response.SmartError(err).Render(w, r)
		}

		cleanupTarget, err := request.ParseCleanupTarget(r.URL.Query())
		if err != nil {
			return response.SmartError(err).Render(w, r)
		}
		if !cleanupTarget {
			return forwardToRemoteClusterConnector(rc, w, r, clusterName, clusterLinksConnectorPath(clusterName, clusterLinkName))
		}

		clusters, err := loadRemoteClusters(rc, r)
		if err != nil {
			return response.SmartError(err).Render(w, r)
		}

		err = deleteClusterLinkWithTargetCleanup(rc, r, clusters, clusterName, clusterLinkName)
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

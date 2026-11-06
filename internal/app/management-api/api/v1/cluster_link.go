package v1

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"time"

	"github.com/canonical/lxd/lxd/response"
	"github.com/canonical/lxd/shared/api"
	"github.com/canonical/microcloud-cluster-manager/internal/app/management-api/core/auth"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/api/models/v1"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/database/store"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/logger"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/types"
	"github.com/jmoiron/sqlx"
	"golang.org/x/sync/errgroup"
)

// ClusterLink is the endpoint group for the cluster links of all remote clusters.
var ClusterLink = types.RouteGroup{
	Prefix: "cluster-link",
	Middlewares: []types.RouteMiddleware{
		auth.AuthMiddleware,
	},
	Endpoints: []types.Endpoint{
		{
			Method:  http.MethodGet,
			Handler: clusterLinksGet,
		},
	},
}

// clusterLinksGet returns the cluster links of all remote clusters with an active tunnel, grouped by remote cluster.
// A cluster whose links cannot be loaded is reported in its group instead of failing the request.
func clusterLinksGet(rc types.RouteConfig) types.EndpointHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		var dbClusters []store.RemoteClusterWithDetail
		err := rc.DB.Transaction(r.Context(), func(ctx context.Context, tx *sqlx.Tx) error {
			var err error
			dbClusters, err = store.GetRemoteClustersWithDetails(ctx, tx)
			return err
		})
		if err != nil {
			return response.SmartError(err).Render(w, r)
		}

		sort.Slice(dbClusters, func(i, j int) bool {
			return dbClusters[i].Name < dbClusters[j].Name
		})

		names := make([]string, 0, len(dbClusters))
		for _, cluster := range dbClusters {
			if cluster.TunnelManagerMemberURL != "" {
				names = append(names, cluster.Name)
			}
		}
		results := CollectClusterLinks(r.Context(), names, clusterLinksTimeout, func(ctx context.Context, name string) ([]api.ClusterLink, error) {
			request := r.Clone(ctx)
			request.Body = http.NoBody
			upstream, err := forwardToClusterConnector(rc, request, name, "/cluster-links")
			if err != nil {
				return nil, err
			}
			defer func() {
				_ = upstream.Body.Close()
			}()
			var lxdResponse struct {
				Metadata []api.ClusterLink `json:"metadata"`
				Error    string            `json:"error"`
			}
			err = json.NewDecoder(upstream.Body).Decode(&lxdResponse)
			if upstream.StatusCode < http.StatusOK || upstream.StatusCode >= http.StatusMultipleChoices {
				if lxdResponse.Error == "" {
					lxdResponse.Error = http.StatusText(upstream.StatusCode)
				}
				return nil, errors.New(lxdResponse.Error)
			}
			if err != nil {
				return nil, fmt.Errorf("failed to parse cluster links response: %w", err)
			}
			return lxdResponse.Metadata, nil
		})
		return response.SyncResponse(true, results).Render(w, r)
	}
}

const (
	// clusterLinksTimeout bounds the time spent on one cluster so that a hung tunnel cannot stall the endpoint.
	clusterLinksTimeout = 5 * time.Second
	// clusterLinksConcurrency is the number of clusters queried at the same time.
	clusterLinksConcurrency = 10
)

// CollectClusterLinks fetches the links of the named clusters concurrently, giving each fetch the timeout.
// The groups are returned in the order of the names.
func CollectClusterLinks(ctx context.Context, names []string, timeout time.Duration, fetch func(context.Context, string) ([]api.ClusterLink, error)) []models.ClusterLinkGroup {
	results := make([]models.ClusterLinkGroup, len(names))
	var workers errgroup.Group
	workers.SetLimit(clusterLinksConcurrency)
	for index, name := range names {
		workers.Go(func() error {
			clusterCtx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()
			group := models.ClusterLinkGroup{Cluster: name, Links: []api.ClusterLink{}}
			links, err := fetch(clusterCtx, name)
			if err != nil {
				group.Error = err.Error()
				logger.Log.Warnw("Failed to load cluster links", "cluster", name, "error", err)
			} else if links != nil {
				group.Links = links
			}
			results[index] = group
			return nil
		})
	}
	_ = workers.Wait()
	return results
}

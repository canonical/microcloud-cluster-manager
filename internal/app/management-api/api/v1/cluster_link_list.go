package v1

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"sort"

	"github.com/canonical/lxd/shared/api"
	"github.com/canonical/microcloud-cluster-manager/internal/app/management-api/core/cluster_link"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/api/models/v1"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/database/store"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/logger"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/types"
	"golang.org/x/sync/errgroup"
)

// listClusterLinks lists the cluster links of the named managed cluster. Only the tunnel of that cluster is queried.
// Targets are resolved against all managed clusters. Returns a not found error for an unknown cluster and a bad
// request error for a cluster without an active tunnel.
func listClusterLinks(rc types.RouteConfig, r *http.Request, clusters []store.RemoteClusterWithDetail, clusterName string) (models.ClusterLinkList, error) {
	source := cluster_link.FindCluster(clusters, clusterName)
	if source == nil {
		return models.ClusterLinkList{}, api.StatusErrorf(http.StatusNotFound, "Remote cluster %q not found", clusterName)
	}

	if source.TunnelManagerMemberURL == "" {
		return models.ClusterLinkList{}, api.StatusErrorf(http.StatusBadRequest, "Remote cluster %q does not have an active tunnel", clusterName)
	}

	list := models.ClusterLinkList{
		ClusterLinks: []models.ResolvedClusterLink{},
		Errors:       []models.ClusterLinkError{},
	}
	links, err := fetchClusterLinks(rc, r, *source)
	if err != nil {
		list.Errors = append(list.Errors, models.ClusterLinkError{Cluster: source.Name, Error: clusterLinkErrorMessage(err)})
	} else {
		list.ClusterLinks = cluster_link.ResolveLinks(source.Name, links, cluster_link.NewTargetIndex(clusters))
	}
	return list, nil
}

// listAllClusterLinks lists the cluster links of all managed clusters with an active tunnel, concurrently and within
// the bounds of the options. Clusters without a tunnel cannot be queried and are skipped. Clusters whose links cannot
// be loaded are reported in the errors of the list, next to the links of the other clusters. Targets are resolved
// against all managed clusters. Entries and errors are ordered by cluster name.
func listAllClusterLinks(rc types.RouteConfig, r *http.Request, clusters []store.RemoteClusterWithDetail, options cluster_link.ListOptions) models.ClusterLinkList {
	ctx := r.Context()
	sources := make([]store.RemoteClusterWithDetail, 0, len(clusters))
	for _, cluster := range clusters {
		if cluster.TunnelManagerMemberURL != "" {
			sources = append(sources, cluster)
		}
	}

	sort.SliceStable(sources, func(i, j int) bool {
		return sources[i].Name < sources[j].Name
	})

	if options.Deadline > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, options.Deadline)
		defer cancel()
	}

	type result struct {
		links   []api.ClusterLink
		err     error
		queried bool
	}

	results := make([]result, len(sources))
	var workers errgroup.Group
	if options.Concurrency > 0 {
		workers.SetLimit(options.Concurrency)
	}

	for index, source := range sources {
		if ctx.Err() != nil {
			break
		}
		workers.Go(func() error {
			clusterCtx := ctx
			if options.ClusterTimeout > 0 {
				var cancel context.CancelFunc
				clusterCtx, cancel = context.WithTimeout(ctx, options.ClusterTimeout)
				defer cancel()
			}

			// Clusters still waiting for a worker when the deadline passed are not queried at all.
			if clusterCtx.Err() != nil {
				return nil
			}
			links, err := fetchClusterLinks(rc, r.WithContext(clusterCtx), source)

			if err != nil {
				logger.Log.Warnw("Failed to load cluster links", "cluster", source.Name, "error", err)
			}

			results[index] = result{links: links, err: err, queried: true}
			return nil
		})
	}

	_ = workers.Wait()

	list := models.ClusterLinkList{
		ClusterLinks: []models.ResolvedClusterLink{},
		Errors:       []models.ClusterLinkError{},
	}

	targets := cluster_link.NewTargetIndex(clusters)
	for index, source := range sources {
		if !results[index].queried {
			continue
		}
		if results[index].err != nil {
			list.Errors = append(list.Errors, models.ClusterLinkError{
				Cluster: source.Name,
				Error:   clusterLinkErrorMessage(results[index].err),
			})

			continue
		}

		list.ClusterLinks = append(list.ClusterLinks, cluster_link.ResolveLinks(source.Name, results[index].links, targets)...)
	}

	return list
}

// fetchClusterLinks loads the cluster links defined on a managed cluster through the cluster connector holding its
// tunnel, on behalf of the user of the request.
func fetchClusterLinks(rc types.RouteConfig, r *http.Request, cluster store.RemoteClusterWithDetail) ([]api.ClusterLink, error) {
	links := []api.ClusterLink{}
	err := requestClusterMember(rc, r, http.MethodGet, cluster, clusterLinksConnectorPath(cluster.Name, ""), &links, cluster_link.MaxResponseBytes)
	if err != nil {
		return nil, err
	}
	if links == nil {
		return nil, api.StatusErrorf(http.StatusBadGateway, "Invalid cluster link list from cluster %q", cluster.Name)
	}

	return links, nil
}

// clusterLinkErrorMessage returns the message of a cluster link list error for the user.
func clusterLinkErrorMessage(err error) string {
	// A request error repeats the URL of the cluster connector member, which is of no use to the user.
	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		err = urlErr.Err
	}

	return err.Error()
}

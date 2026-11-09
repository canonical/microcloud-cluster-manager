package v1

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/canonical/lxd/shared/api"
	"github.com/canonical/microcloud-cluster-manager/internal/app/management-api/core/connector"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/api/models/v1"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/database/store"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/logger"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/request"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/types"
	"golang.org/x/sync/errgroup"
)

// addClusterLinks sets the cluster links of the remote clusters with an active tunnel. A cluster whose links cannot be
// loaded is reported in its own error field instead of failing the request. The db clusters must hold the clusters
// that are listed.
func addClusterLinks(rc types.RouteConfig, r *http.Request, dbClusters []store.RemoteClusterWithDetail, clusters []models.RemoteCluster) {
	// Clusters without a tunnel cannot be queried and have no links to report.
	tunneledClusters := make([]store.RemoteClusterWithDetail, 0, len(dbClusters))
	for _, cluster := range dbClusters {
		if cluster.TunnelManagerMemberURL != "" {
			tunneledClusters = append(tunneledClusters, cluster)
		}
	}

	results := collectClusterLinks(rc, r, tunneledClusters)
	for i := range clusters {
		result, ok := results[clusters[i].Name]
		if !ok {
			continue
		}

		clusters[i].ClusterLinks = result.links
		clusters[i].ClusterLinksError = result.errMessage
	}
}

// clusterLinksResult holds the cluster links of one remote cluster, or why they could not be loaded.
type clusterLinksResult struct {
	links      []api.ClusterLink
	errMessage string
}

// clusterLinksConnectorPath is the cluster connector path of the cluster links of a remote cluster.
const clusterLinksConnectorPath = "/cluster-links"

const (
	// clusterLinksTimeout bounds the time spent on one cluster so that a hung tunnel cannot stall the endpoint.
	clusterLinksTimeout = 10 * time.Second
	// clusterLinksTotalTimeout bounds the time spent on all clusters together.
	clusterLinksTotalTimeout = 15 * time.Second
	// clusterLinksResponseMargin is the time kept free for writing the response before the server write timeout.
	clusterLinksResponseMargin = time.Second
	// maxClusterLinksResponseBytes limits the size of the cluster links of one cluster.
	maxClusterLinksResponseBytes = 4 << 20 // 4 MiB
	// clusterLinksConcurrencyLimit is the max number of clusters queried at the same time.
	clusterLinksConcurrencyLimit = 40
)

// clusterLinksDeadline returns how long the cluster links of all clusters may take together. Clusters are queried in
// batches, so without it enough slow clusters would make the request outlast the server and the client would get no
// response at all instead of the links that could be loaded.
func clusterLinksDeadline(rc types.RouteConfig) time.Duration {
	deadline := clusterLinksTotalTimeout

	// The server closes the connection once its write timeout passed, so the response must be ready before that.
	writeTimeout := time.Duration(rc.Env.WriteTimeout) * time.Second
	if writeTimeout > 0 {
		deadline = min(deadline, max(writeTimeout-clusterLinksResponseMargin, clusterLinksTimeout))
	}

	return deadline
}

// collectClusterLinks loads the cluster links of the remote clusters through the cluster connector, concurrently and
// giving each cluster its own timeout. Clusters that are not done when the overall deadline passes are reported with
// the resulting error in their result. The results are returned by cluster name.
func collectClusterLinks(rc types.RouteConfig, r *http.Request, clusters []store.RemoteClusterWithDetail) map[string]clusterLinksResult {
	ctx := r.Context()
	deadline := clusterLinksDeadline(rc)
	if deadline > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, deadline)
		defer cancel()
	}

	results := make([]clusterLinksResult, len(clusters))
	var workers errgroup.Group
	workers.SetLimit(clusterLinksConcurrencyLimit)
	for index, cluster := range clusters {
		name := cluster.Name
		workers.Go(func() error {
			result := clusterLinksResult{links: []api.ClusterLink{}}
			defer func() {
				results[index] = result
			}()

			fail := func(err error) error {
				result.errMessage = err.Error()
				logger.Log.Warnw("Failed to load cluster links", "cluster", name, "error", err)
				return nil
			}

			clusterCtx, cancel := context.WithTimeout(ctx, clusterLinksTimeout)
			defer cancel()

			clusterReq := r.Clone(clusterCtx)
			clusterReq.Body = http.NoBody
			upstream, err := connector.Forward(rc, clusterReq, name, clusterLinksConnectorPath, cluster.TunnelManagerMemberURL)
			if err != nil {
				// The request error repeats the URL of the cluster connector member, which is of no use to the user.
				var urlErr *url.Error
				if errors.As(err, &urlErr) {
					err = urlErr.Err
				}

				return fail(err)
			}
			defer func() {
				_ = upstream.Body.Close()
			}()

			// Read one byte more than allowed to tell a response that is exactly at the limit from a larger one.
			body, decodeErr := io.ReadAll(io.LimitReader(upstream.Body, maxClusterLinksResponseBytes+1))
			if len(body) > maxClusterLinksResponseBytes {
				return fail(fmt.Errorf("cluster links response is larger than %d bytes", maxClusterLinksResponseBytes))
			}

			var lxdResponse struct {
				Metadata []api.ClusterLink `json:"metadata"`
				Error    string            `json:"error"`
			}
			if decodeErr == nil {
				decodeErr = json.Unmarshal(body, &lxdResponse)
			}

			if !request.IsSuccessStatus(upstream.StatusCode) {
				if lxdResponse.Error == "" {
					lxdResponse.Error = http.StatusText(upstream.StatusCode)
				}

				return fail(errors.New(lxdResponse.Error))
			}

			if decodeErr != nil {
				return fail(fmt.Errorf("failed to parse cluster links response: %w", decodeErr))
			}

			if lxdResponse.Metadata != nil {
				result.links = lxdResponse.Metadata
			}

			return nil
		})
	}

	_ = workers.Wait()

	resultsByCluster := make(map[string]clusterLinksResult, len(clusters))
	for index, cluster := range clusters {
		resultsByCluster[cluster.Name] = results[index]
	}

	return resultsByCluster
}

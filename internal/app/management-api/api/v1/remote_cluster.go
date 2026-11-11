package v1

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/canonical/lxd/lxd/response"
	"github.com/canonical/lxd/shared"
	"github.com/canonical/lxd/shared/api"
	"github.com/canonical/microcloud-cluster-manager/internal/app/management-api/core/auth"
	"github.com/canonical/microcloud-cluster-manager/internal/app/management-api/core/cluster_link"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/api/models/v1"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/database/store"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/logger"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/request"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/types"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
)

// RemoteCluster is the remote cluster endpoint group.
var RemoteCluster = types.RouteGroup{
	Prefix: "remote-cluster",
	Middlewares: []types.RouteMiddleware{
		auth.AuthMiddleware,
	},
	Endpoints: []types.Endpoint{
		{
			Method:  http.MethodGet,
			Handler: remoteClustersGet,
		},
		{
			Path:    "{remoteClusterName}",
			Method:  http.MethodGet,
			Handler: remoteClusterGet,
		},
		{
			Path:    "{remoteClusterName}",
			Method:  http.MethodDelete,
			Handler: remoteClusterDelete,
		},
		{
			Path:    "{remoteClusterName}",
			Method:  http.MethodPatch,
			Handler: remoteClusterPatch,
		},
		{
			Path:    "{remoteClusterName}/auth-groups",
			Method:  http.MethodGet,
			Handler: remoteClusterAuthGroupsGet,
		},
		{
			Path:    "{remoteClusterName}/cluster-links",
			Method:  http.MethodGet,
			Handler: remoteClusterLinksGet,
		},
		{
			Path:    "{remoteClusterName}/cluster-links",
			Method:  http.MethodPost,
			Handler: remoteClusterLinksPost,
		},
		{
			Path:    "{remoteClusterName}/cluster-links/{clusterLinkName}",
			Method:  http.MethodPut,
			Handler: remoteClusterLinkPut,
		},
		{
			Path:    "{remoteClusterName}/cluster-links/{clusterLinkName}",
			Method:  http.MethodGet,
			Handler: remoteClusterLinkGet,
		},
		{
			Path:    "{remoteClusterName}/cluster-links/{clusterLinkName}/state",
			Method:  http.MethodGet,
			Handler: remoteClusterLinkStateGet,
		},
		{
			Path:    "{remoteClusterName}/cluster-links/{clusterLinkName}",
			Method:  http.MethodDelete,
			Handler: remoteClusterLinkDelete,
		},
		{
			Path:    "{remoteClusterName}/replicators",
			Method:  http.MethodGet,
			Handler: remoteClusterReplicatorsGet,
		},
		{
			Path:    "{remoteClusterName}/replicators",
			Method:  http.MethodPost,
			Handler: remoteClusterReplicatorsPost,
		},
		{
			Path:    "{remoteClusterName}/replicators/{replicatorName}",
			Method:  http.MethodGet,
			Handler: remoteClusterReplicatorGet,
		},
		{
			Path:    "{remoteClusterName}/replicators/{replicatorName}",
			Method:  http.MethodPut,
			Handler: remoteClusterReplicatorPut,
		},
		{
			Path:    "{remoteClusterName}/replicators/{replicatorName}",
			Method:  http.MethodDelete,
			Handler: remoteClusterReplicatorDelete,
		},
		{
			Path:    "{remoteClusterName}/replicators/{replicatorName}/state",
			Method:  http.MethodGet,
			Handler: remoteClusterReplicatorStateGet,
		},
		{
			Path:    "{remoteClusterName}/replicators/{replicatorName}/state",
			Method:  http.MethodPut,
			Handler: remoteClusterReplicatorStatePut,
		},
		{
			Path:    "{remoteClusterName}/projects",
			Method:  http.MethodGet,
			Handler: remoteClusterProjectsGet,
		},
		{
			Path:    "{remoteClusterName}/projects",
			Method:  http.MethodPost,
			Handler: remoteClusterProjectsPost,
		},
		{
			Path:    "{remoteClusterName}/projects/{projectName}",
			Method:  http.MethodGet,
			Handler: remoteClusterProjectGet,
		},
		{
			Path:    "{remoteClusterName}/projects/{projectName}",
			Method:  http.MethodPut,
			Handler: remoteClusterProjectPut,
		},
		{
			Path:    "{remoteClusterName}/projects/{projectName}/state",
			Method:  http.MethodGet,
			Handler: remoteClusterProjectStateGet,
		},
		{
			Path:    "{remoteClusterName}/projects/{projectName}/state",
			Method:  http.MethodPut,
			Handler: remoteClusterProjectStatePut,
		},
		{
			Path:    "{remoteClusterName}/operations/{operationID}",
			Method:  http.MethodGet,
			Handler: remoteClusterOperationGet,
		},
		{
			Path:    "{remoteClusterName}/auth/identities",
			Method:  http.MethodGet,
			Handler: remoteClusterIdentitiesGet,
		},
	},
}

func remoteClustersGet(rc types.RouteConfig) types.EndpointHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		var dbRemoteClusterDetails []store.RemoteClusterWithDetail

		err := rc.DB.Transaction(r.Context(), func(ctx context.Context, tx *sqlx.Tx) error {
			var err error
			dbRemoteClusterDetails, err = store.GetRemoteClustersWithDetails(ctx, tx)
			return err
		})

		if err != nil {
			return response.SmartError(err).Render(w, r)
		}

		result, err := toRemoteClustersAPI(dbRemoteClusterDetails)
		if err != nil {
			return response.InternalError(err).Render(w, r)
		}

		return response.SyncResponse(true, result).Render(w, r)
	}
}

func remoteClusterGet(rc types.RouteConfig) types.EndpointHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		remoteClusterName, err := request.RemoteClusterName(r)
		if err != nil {
			return response.SmartError(err).Render(w, r)
		}

		var dbRemoteClusterDetail *store.RemoteClusterWithDetail
		err = rc.DB.Transaction(r.Context(), func(ctx context.Context, tx *sqlx.Tx) error {
			var err error
			dbRemoteClusterDetail, err = store.GetRemoteClusterWithDetailByName(ctx, tx, remoteClusterName)
			return err
		})

		if err != nil {
			return response.SmartError(err).Render(w, r)
		}

		if dbRemoteClusterDetail == nil {
			return response.NotFound(fmt.Errorf("RemoteCluster not found")).Render(w, r)
		}

		result, err := toRemoteClustersAPI([]store.RemoteClusterWithDetail{*dbRemoteClusterDetail})
		if err != nil {
			return response.InternalError(err).Render(w, r)
		}

		return response.SyncResponse(true, result[0]).Render(w, r)
	}
}

func remoteClusterDelete(rc types.RouteConfig) types.EndpointHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		remoteClusterName, err := request.RemoteClusterName(r)
		if err != nil {
			return response.SmartError(err).Render(w, r)
		}

		err = rc.DB.Transaction(r.Context(), func(ctx context.Context, tx *sqlx.Tx) error {
			return store.DeleteRemoteCluster(ctx, tx, remoteClusterName)
		})

		if err != nil {
			return response.SmartError(err).Render(w, r)
		}

		return response.EmptySyncResponse.Render(w, r)
	}
}

func remoteClusterPatch(rc types.RouteConfig) types.EndpointHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		remoteClusterName, err := request.RemoteClusterName(r)
		if err != nil {
			return response.SmartError(err).Render(w, r)
		}

		var payload models.RemoteClusterPatch
		err = json.NewDecoder(r.Body).Decode(&payload)
		if err != nil {
			return response.BadRequest(err).Render(w, r)
		}

		if payload.Status != "" {
			if payload.Status != models.ACTIVE {
				return response.BadRequest(fmt.Errorf("invalid status")).Render(w, r)
			}
		}

		err = rc.DB.Transaction(r.Context(), func(ctx context.Context, tx *sqlx.Tx) error {
			existingRemoteCluster, err := store.GetRemoteCluster(ctx, tx, remoteClusterName)
			if err != nil {
				return err
			}

			newRemoteCluster := existingRemoteCluster
			if payload.Status != "" {
				newRemoteCluster.Status = string(payload.Status)
			}
			if payload.Description != nil {
				newRemoteCluster.Description = *payload.Description
			}

			if payload.DiskThreshold < 0 || payload.DiskThreshold > 100 {
				return errors.New("disk threshold outside 0 and 100 percent")
			}
			if payload.DiskThreshold > 0 {
				err = store.UpdateRemoteClusterConfig(ctx, tx, store.RemoteClusterConfig{
					RemoteClusterID: existingRemoteCluster.ID,
					Key:             store.DiskThresholdKey,
					Value:           fmt.Sprintf("%d", payload.DiskThreshold),
				})
				if err != nil {
					return err
				}
			}

			if payload.MemoryThreshold < 0 || payload.MemoryThreshold > 100 {
				return errors.New("memory threshold outside 0 and 100 percent")
			}
			if payload.MemoryThreshold > 0 {
				err = store.UpdateRemoteClusterConfig(ctx, tx, store.RemoteClusterConfig{
					RemoteClusterID: existingRemoteCluster.ID,
					Key:             store.MemoryThresholdKey,
					Value:           fmt.Sprintf("%d", payload.MemoryThreshold),
				})
				if err != nil {
					return err
				}
			}

			newRemoteCluster.UpdatedAt = time.Now()

			return store.UpdateRemoteCluster(ctx, tx, remoteClusterName, *newRemoteCluster)
		})

		if err != nil {
			return response.SmartError(err).Render(w, r)
		}

		return response.EmptySyncResponse.Render(w, r)
	}
}

func toRemoteClustersAPI(dbEntries []store.RemoteClusterWithDetail) ([]models.RemoteCluster, error) {
	// generate lookup for remoteCluster details
	var remoteClusters []models.RemoteCluster
	for _, e := range dbEntries {
		var cephStatuses []models.StatusDistribution
		err := json.Unmarshal(e.CephStatuses, &cephStatuses)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal ceph statuses: %w", err)
		}

		var memberStatuses []models.StatusDistribution
		err = json.Unmarshal(e.MemberStatuses, &memberStatuses)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal member statuses: %w", err)
		}

		var instanceStatuses []models.StatusDistribution
		err = json.Unmarshal(e.InstanceStatuses, &instanceStatuses)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal instance statuses: %w", err)
		}

		var storagePoolUsages []models.StoragePoolUsage
		err = json.Unmarshal(e.StoragePoolUsages, &storagePoolUsages)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal storage pool usages: %w", err)
		}

		remoteClusters = append(remoteClusters, models.RemoteCluster{
			Name:               e.Name,
			ClusterUUID:        e.ClusterUUID,
			Description:        e.Description,
			ClusterCertificate: e.ClusterCertificate,
			DiskThreshold:      e.DiskThreshold,
			MemoryThreshold:    e.MemoryThreshold,
			Status:             e.Status,
			CephCount:          e.CephCount,
			CephStatuses:       cephStatuses,
			CPUTotalCount:      e.CPUTotalCount,
			CPULoad1:           e.CPULoad1,
			CPULoad5:           e.CPULoad5,
			CPULoad15:          e.CPULoad15,
			MemoryTotalAmount:  e.MemoryTotalAmount,
			MemoryUsage:        e.MemoryUsage,
			MemberCount:        e.MemberCount,
			MemberStatuses:     memberStatuses,
			InstanceCount:      e.InstanceCount,
			InstanceStatuses:   instanceStatuses,
			StoragePoolUsages:  storagePoolUsages,
			TunnelRegistered:   e.TunnelManagerMemberURL != "",
			LXDURL:             e.LXDURL,
			JoinedAt:           e.ClusterJoinedAt,
			CreatedAt:          e.ClusterCreatedAt,
			LastStatusUpdateAt: e.ClusterUpdatedAt,
		})
	}

	sort.Slice(remoteClusters, func(i, j int) bool {
		return remoteClusters[i].Name < remoteClusters[j].Name
	})

	return remoteClusters, nil
}

func remoteClusterAuthGroupsGet(rc types.RouteConfig) types.EndpointHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		return forwardToClusterConnector(rc, w, r, "/auth-groups")
	}
}

func remoteClusterLinksPost(rc types.RouteConfig) types.EndpointHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		clusterConnectorPath := "/cluster-links"
		return forwardToClusterConnector(rc, w, r, clusterConnectorPath)
	}
}

func remoteClusterLinksGet(rc types.RouteConfig) types.EndpointHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		clusterConnectorPath := "/cluster-links"
		return forwardToClusterConnector(rc, w, r, clusterConnectorPath)
	}
}

func remoteClusterLinkGet(rc types.RouteConfig) types.EndpointHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		clusterLinkName, err := request.PathParam(r, request.ClusterLinkNameParam, "cluster link name")
		if err != nil {
			return response.SmartError(err).Render(w, r)
		}

		clusterConnectorPath := "/cluster-links/" + url.PathEscape(clusterLinkName)
		return forwardToClusterConnector(rc, w, r, clusterConnectorPath)
	}
}

func remoteClusterLinkStateGet(rc types.RouteConfig) types.EndpointHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		clusterLinkName, err := url.PathUnescape(mux.Vars(r)["clusterLinkName"])
		if err != nil {
			return response.SmartError(err).Render(w, r)
		}

		clusterConnectorPath := "/cluster-links/" + url.PathEscape(clusterLinkName) + "/state"
		return forwardToClusterConnector(rc, w, r, clusterConnectorPath)
	}
}

func remoteClusterIdentitiesGet(rc types.RouteConfig) types.EndpointHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		clusterConnectorPath := "/auth/identities"
		return forwardToClusterConnector(rc, w, r, clusterConnectorPath)
	}
}

func remoteClusterLinkPut(rc types.RouteConfig) types.EndpointHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		clusterLinkName, err := request.PathParam(r, request.ClusterLinkNameParam, "cluster link name")
		if err != nil {
			return response.SmartError(err).Render(w, r)
		}

		clusterConnectorPath := "/cluster-links/" + url.PathEscape(clusterLinkName)
		return forwardToClusterConnector(rc, w, r, clusterConnectorPath)
	}
}

// remoteClusterLinkDelete deletes a cluster link on the given remote cluster. When the removeReverseEdge query
// parameter is true, it also removes the other edge of the link on the linked cluster: the reverse cluster link for
// bidirectional links, or the cluster link identity for unidirectional links. The link type and the linked cluster
// are resolved from the link on the source cluster.
func remoteClusterLinkDelete(rc types.RouteConfig) types.EndpointHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		clusterLinkName, err := request.PathParam(r, request.ClusterLinkNameParam, "cluster link name")
		if err != nil {
			return response.SmartError(err).Render(w, r)
		}

		clusterConnectorPath := "/cluster-links/" + url.PathEscape(clusterLinkName)

		removeReverseEdge := shared.IsTrue(r.URL.Query().Get("removeReverseEdge"))
		if !removeReverseEdge {
			return forwardToClusterConnector(rc, w, r, clusterConnectorPath)
		}

		sourceClusterName, err := url.PathUnescape(mux.Vars(r)["remoteClusterName"])
		if err != nil {
			return response.SmartError(err).Render(w, r)
		}

		var sourceLinks []api.ClusterLink
		err = requestClusterConnector(rc, r, http.MethodGet, sourceClusterName, "/cluster-links", &sourceLinks)
		if err != nil {
			return response.SmartError(err).Render(w, r)
		}

		sourceLink := cluster_link.FindLinkByName(sourceLinks, clusterLinkName)
		if sourceLink == nil {
			return response.NotFound(fmt.Errorf("cluster link %q not found on cluster %q", clusterLinkName, sourceClusterName)).Render(w, r)
		}

		linkType := strings.ToLower(sourceLink.Type)
		if linkType != api.ClusterLinkTypeBidirectional && linkType != api.ClusterLinkTypeUnidirectional {
			return response.BadRequest(fmt.Errorf("removing the reverse edge is only supported for %s and %s cluster links", api.ClusterLinkTypeBidirectional, api.ClusterLinkTypeUnidirectional)).Render(w, r)
		}

		var clusters []store.RemoteClusterWithDetail
		err = rc.DB.Transaction(r.Context(), func(ctx context.Context, tx *sqlx.Tx) error {
			var err error
			clusters, err = store.GetRemoteClustersWithDetails(ctx, tx)
			return err
		})
		if err != nil {
			return response.SmartError(err).Render(w, r)
		}

		var sourceCluster *store.RemoteClusterWithDetail
		for i := range clusters {
			if clusters[i].Name == sourceClusterName {
				sourceCluster = &clusters[i]
				break
			}
		}

		if sourceCluster == nil {
			return response.NotFound(fmt.Errorf("remote cluster %q not found", sourceClusterName)).Render(w, r)
		}

		targetCluster := findLinkedCluster(*sourceLink, clusters, sourceClusterName)
		if targetCluster == nil {
			return response.BadRequest(fmt.Errorf("cluster link %q on cluster %q does not point at a registered cluster", clusterLinkName, sourceClusterName)).Render(w, r)
		}

		targetClusterName := targetCluster.Name

		// Remove the target edge before the source link. A failure on the target then leaves the source link in
		// place, so the deletion can be retried. Target cleanup treats an already missing edge as done, so a retry
		// after the target edge was removed but the source link deletion failed also succeeds.
		if linkType == api.ClusterLinkTypeBidirectional {
			err = deleteReverseClusterLink(rc, r, sourceCluster, targetClusterName)
		} else {
			err = deleteClusterLinkIdentity(rc, r, sourceClusterName, targetClusterName)
		}

		if err != nil {
			logger.Log.Errorw("Failed to delete the target edge of a cluster link", "source_cluster", sourceClusterName, "target_cluster", targetClusterName, "cluster_link", clusterLinkName, "type", linkType, "error", err)
			return response.SmartError(fmt.Errorf("failed to delete the link edge on cluster %q, cluster link %q on cluster %q was not deleted: %w", targetClusterName, clusterLinkName, sourceClusterName, err)).Render(w, r)
		}

		err = requestClusterConnector(rc, r, http.MethodDelete, sourceClusterName, clusterConnectorPath, nil)
		if err != nil {
			return response.SmartError(err).Render(w, r)
		}

		return response.EmptySyncResponse.Render(w, r)
	}
}

// findLinkedCluster resolves the registered cluster the link points at, other than the source cluster. A match by
// cluster UUID is preferred over a match by address.
func findLinkedCluster(link api.ClusterLink, clusters []store.RemoteClusterWithDetail, sourceClusterName string) *store.RemoteClusterWithDetail {
	for i := range clusters {
		if clusters[i].Name != sourceClusterName && cluster_link.LinkPointsAt(link, clusters[i].ClusterUUID, "") {
			return &clusters[i]
		}
	}

	for i := range clusters {
		if clusters[i].Name != sourceClusterName && cluster_link.LinkPointsAt(link, "", clusters[i].LXDURL) {
			return &clusters[i]
		}
	}

	return nil
}

// deleteReverseClusterLink deletes the bidirectional cluster link on the target cluster that points back at the
// source cluster. A missing reverse link is treated as already deleted.
func deleteReverseClusterLink(rc types.RouteConfig, r *http.Request, sourceCluster *store.RemoteClusterWithDetail, targetClusterName string) error {
	var targetLinks []api.ClusterLink
	err := requestClusterConnector(rc, r, http.MethodGet, targetClusterName, "/cluster-links", &targetLinks)
	if err != nil {
		return err
	}

	reverseLink := cluster_link.FindReverseLink(targetLinks, sourceCluster.ClusterUUID, sourceCluster.LXDURL)
	if reverseLink == nil {
		logger.Log.Infow("No reverse cluster link found on target cluster, assuming it was already deleted", "source_cluster", sourceCluster.Name, "target_cluster", targetClusterName)
		return nil
	}

	err = requestClusterConnector(rc, r, http.MethodDelete, targetClusterName, "/cluster-links/"+url.PathEscape(reverseLink.Name), nil)
	if api.StatusErrorCheck(err, http.StatusNotFound) {
		return nil
	}

	return err
}

// deleteClusterLinkIdentity deletes the cluster link identity on the target cluster that holds the certificate of
// the source cluster. A missing identity is treated as already deleted.
func deleteClusterLinkIdentity(rc types.RouteConfig, r *http.Request, sourceClusterName string, targetClusterName string) error {
	var sourceServer api.Server
	err := requestClusterConnector(rc, r, http.MethodGet, sourceClusterName, "/info", &sourceServer)
	if err != nil {
		return err
	}

	fingerprint := sourceServer.Environment.CertificateFingerprint
	if fingerprint == "" && sourceServer.Environment.Certificate != "" {
		fingerprint, err = shared.CertFingerprintStr(sourceServer.Environment.Certificate)
		if err != nil {
			return api.StatusErrorf(http.StatusBadGateway, "cluster %q reported an invalid certificate: %v", sourceClusterName, err)
		}
	}

	if fingerprint == "" {
		return api.StatusErrorf(http.StatusBadGateway, "cluster %q did not report a certificate", sourceClusterName)
	}

	var identities []api.Identity
	err = requestClusterConnector(rc, r, http.MethodGet, targetClusterName, "/auth/identities", &identities)
	if err != nil {
		return err
	}

	identity := cluster_link.FindIdentity(identities, fingerprint)
	if identity == nil {
		logger.Log.Infow("No cluster link identity found on target cluster, assuming it was already deleted", "source_cluster", sourceClusterName, "target_cluster", targetClusterName)
		return nil
	}

	err = requestClusterConnector(rc, r, http.MethodDelete, targetClusterName, "/auth/identities/tls/"+url.PathEscape(identity.Identifier), nil)
	if api.StatusErrorCheck(err, http.StatusNotFound) {
		return nil
	}

	return err
}

// requestClusterConnector sends a request without a body for the given remote cluster to the cluster connector and
// decodes the metadata of the response into target, unless target is nil. A non-2xx response is returned as an error
// carrying the upstream status code.
func requestClusterConnector(rc types.RouteConfig, r *http.Request, method string, remoteClusterName string, clusterConnectorPath string, target any) error {
	resp, err := sendToClusterConnector(rc, r, method, remoteClusterName, clusterConnectorPath, nil)
	if err != nil {
		return err
	}

	defer func() {
		err := resp.Body.Close()
		if err != nil {
			logger.Log.Errorf("failed to close response body: %v", err)
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response for cluster %q: %w", remoteClusterName, err)
	}

	var lxdResponse struct {
		Metadata json.RawMessage `json:"metadata"`
		Error    string          `json:"error"`
	}

	decodeErr := json.Unmarshal(body, &lxdResponse)

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		if decodeErr != nil || lxdResponse.Error == "" {
			return api.StatusErrorf(resp.StatusCode, "request %s %s for cluster %q failed with HTTP status %d", method, clusterConnectorPath, remoteClusterName, resp.StatusCode)
		}

		return api.StatusErrorf(resp.StatusCode, "%s", lxdResponse.Error)
	}

	if target == nil {
		return nil
	}

	if decodeErr != nil {
		return api.StatusErrorf(http.StatusBadGateway, "invalid response for cluster %q: %v", remoteClusterName, decodeErr)
	}

	err = json.Unmarshal(lxdResponse.Metadata, target)
	if err != nil {
		return api.StatusErrorf(http.StatusBadGateway, "invalid response metadata for cluster %q: %v", remoteClusterName, err)
	}

	return nil
}

func forwardToClusterConnector(rc types.RouteConfig, w http.ResponseWriter, r *http.Request, clusterConnectorPath string) error {
	remoteClusterName, err := request.RemoteClusterName(r)
	if err != nil {
		return response.SmartError(err).Render(w, r)
	}

	// Buffer the body so that it can be sent again if the request needs to be retried after a token refresh.
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return response.SmartError(fmt.Errorf("failed to read request body: %w", err)).Render(w, r)
	}

	defer func() {
		err := r.Body.Close()
		if err != nil {
			logger.Log.Warnw("Failed to close request body", "error", err)
		}
	}()

	resp, err := sendToClusterConnector(rc, r, r.Method, remoteClusterName, clusterConnectorPath, body)
	if err != nil {
		return response.SmartError(err).Render(w, r)
	}

	defer func() {
		err := resp.Body.Close()
		if err != nil {
			logger.Log.Errorf("failed to close response body: %v", err)
		}
	}()

	for k, values := range resp.Header {
		w.Header()[k] = append([]string(nil), values...)
	}

	w.WriteHeader(resp.StatusCode)
	if _, err = io.Copy(w, resp.Body); err != nil {
		logger.Log.Errorf("failed to stream response body", "err", err)
		return nil
	}

	return nil
}

// sendToClusterConnector sends a request for the given remote cluster to the cluster connector holding its tunnel.
// Requests without side effects are retried once with refreshed session tokens after a 401 or 403 response.
// The caller must close the body of the returned response.
func sendToClusterConnector(rc types.RouteConfig, r *http.Request, method string, remoteClusterName string, clusterConnectorPath string, body []byte) (*http.Response, error) {
	var clusterDetails *store.RemoteClusterDetail
	err := rc.DB.Transaction(r.Context(), func(ctx context.Context, tx *sqlx.Tx) error {
		cluster, err := store.GetRemoteCluster(ctx, tx, remoteClusterName)
		if err != nil {
			return err
		}

		clusterDetails, err = store.GetRemoteClusterDetail(ctx, tx, cluster.ID)
		return err
	})
	if err != nil {
		return nil, err
	}

	if clusterDetails.TunnelManagerMemberURL == "" {
		return nil, api.StatusErrorf(http.StatusBadRequest, "remote cluster %q does not have an active tunnel", remoteClusterName)
	}

	reqURL := clusterDetails.TunnelManagerMemberURL + "/1.0/remote-cluster/" + url.PathEscape(remoteClusterName) + clusterConnectorPath

	sessionProvider, ok := r.Context().Value(auth.SessionCredentialsProviderKey).(*auth.SessionCredentialsProvider)
	if !ok {
		return nil, api.StatusErrorf(http.StatusInternalServerError, "failed to get session credentials provider from the request context")
	}

	accessToken, err := sessionProvider.GetAccessToken()
	if err != nil {
		return nil, api.StatusErrorf(http.StatusInternalServerError, "failed to get access token: %v", err)
	}

	userSecret, err := sessionProvider.GetUserSecret()
	if err != nil {
		return nil, api.StatusErrorf(http.StatusInternalServerError, "failed to get user secret: %v", err)
	}

	resp, err := sendInternalRequest(r, method, accessToken, userSecret, reqURL, rc, body)
	if err != nil {
		return nil, err
	}

	// Only replay requests that have no side effects. A 401/403 from a non-idempotent request may be the
	// result of a partially applied change, so replaying it could apply the change twice.
	isRetriable := method == http.MethodGet || method == http.MethodHead || method == http.MethodOptions
	needsTokenRefresh := isRetriable && (resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden)
	if needsTokenRefresh {
		refreshCtx, cancelRefresh := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancelRefresh()
		accessToken, err = sessionProvider.RefreshSessionTokens(refreshCtx)
		if err != nil {
			return nil, api.StatusErrorf(http.StatusInternalServerError, "failed to refresh access token: %v", err)
		}

		err = resp.Body.Close()
		if err != nil {
			logger.Log.Warnw("Failed to close response body before retry", "error", err)
		}

		resp, err = sendInternalRequest(r, method, accessToken, userSecret, reqURL, rc, body)
		if err != nil {
			return nil, api.StatusErrorf(http.StatusInternalServerError, "failed to send request after token refresh: %v", err)
		}
	}

	return resp, nil
}

func sendInternalRequest(r *http.Request, method string, accessToken string, userSecret string, reqURL string, rc types.RouteConfig, body []byte) (*http.Response, error) {
	req, err := http.NewRequestWithContext(r.Context(), method, reqURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	if contentType := r.Header.Get("Content-Type"); contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("X-User-Secret", userSecret)

	// send request
	serverCert := rc.Env.ClusterConnectorCert.CA()
	tlsConfig := shared.InitTLSConfig()
	tlsConfig.Certificates = []tls.Certificate{}
	tlsConfig.RootCAs = x509.NewCertPool()
	tlsConfig.RootCAs.AddCert(serverCert)
	cert := rc.Env.ManagementAPICert.KeyPair()
	tlsConfig.GetClientCertificate = func(info *tls.CertificateRequestInfo) (*tls.Certificate, error) {
		// GetClientCertificate is called if not nil instead of performing the default selection of an appropriate
		// certificate from the `Certificates` list. We only have one-key pair to send, and we always want to send it
		// because this is what uniquely identifies the caller to the server.
		return &cert, nil
	}
	client := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

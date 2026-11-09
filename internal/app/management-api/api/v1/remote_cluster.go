package v1

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"time"

	"github.com/canonical/lxd/lxd/response"
	"github.com/canonical/microcloud-cluster-manager/internal/app/management-api/core/auth"
	"github.com/canonical/microcloud-cluster-manager/internal/app/management-api/core/connector"
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
			Method:  http.MethodDelete,
			Handler: remoteClusterLinkDelete,
		},
	},
}

func remoteClustersGet(rc types.RouteConfig) types.EndpointHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		withLinks, err := parseWithLinks(r)
		if err != nil {
			return response.BadRequest(err).Render(w, r)
		}

		var dbRemoteClusterDetails []store.RemoteClusterWithDetail

		err = rc.DB.Transaction(r.Context(), func(ctx context.Context, tx *sqlx.Tx) error {
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

		if withLinks {
			addClusterLinks(rc, r, dbRemoteClusterDetails, result)
		}

		return response.SyncResponse(true, result).Render(w, r)
	}
}

// parseWithLinks returns whether the with-links query parameter asks for the cluster links of each cluster.
func parseWithLinks(r *http.Request) (bool, error) {
	value := r.URL.Query().Get("with-links")
	if value == "" {
		return false, nil
	}

	withLinks, err := strconv.ParseBool(value)
	if err != nil {
		return false, fmt.Errorf("invalid with-links value %q", value)
	}

	return withLinks, nil
}

func remoteClusterGet(rc types.RouteConfig) types.EndpointHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		remoteClusterName, err := url.PathUnescape(mux.Vars(r)["remoteClusterName"])
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
		remoteClusterName, err := url.PathUnescape(mux.Vars(r)["remoteClusterName"])
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
		remoteClusterName, err := url.PathUnescape(mux.Vars(r)["remoteClusterName"])
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

// authGroupsConnectorPath is the cluster connector path of the auth groups of a remote cluster.
const authGroupsConnectorPath = "/auth-groups"

func remoteClusterAuthGroupsGet(rc types.RouteConfig) types.EndpointHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		return forwardToClusterConnector(rc, w, r, authGroupsConnectorPath)
	}
}

func remoteClusterLinksGet(rc types.RouteConfig) types.EndpointHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		return forwardToClusterConnector(rc, w, r, clusterLinksConnectorPath)
	}
}

func remoteClusterLinksPost(rc types.RouteConfig) types.EndpointHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		return forwardToClusterConnector(rc, w, r, clusterLinksConnectorPath)
	}
}

func remoteClusterLinkDelete(rc types.RouteConfig) types.EndpointHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		clusterLinkName, err := url.PathUnescape(mux.Vars(r)["clusterLinkName"])
		if err != nil {
			return response.SmartError(err).Render(w, r)
		}

		return forwardToClusterConnector(rc, w, r, clusterLinksConnectorPath+"/"+url.PathEscape(clusterLinkName))
	}
}

// forwardToClusterConnector forwards the request to the cluster connector path of the remote cluster of the route and
// streams the response back.
func forwardToClusterConnector(rc types.RouteConfig, w http.ResponseWriter, r *http.Request, clusterConnectorPath string) error {
	remoteClusterName, err := url.PathUnescape(mux.Vars(r)["remoteClusterName"])
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

	resp, err := connector.Forward(rc, r, remoteClusterName, clusterConnectorPath, dbRemoteClusterDetail.TunnelManagerMemberURL)
	if err != nil {
		return response.SmartError(err).Render(w, r)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	request.CopyResponseHeaders(w.Header(), resp.Header)
	w.WriteHeader(resp.StatusCode)
	_, err = io.Copy(w, resp.Body)

	if err != nil {
		logger.Log.Errorw("Failed to stream response body", "error", err)
	}
	return nil
}

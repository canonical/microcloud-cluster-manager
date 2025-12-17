package v1

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"time"

	"github.com/canonical/lxd/lxd/response"
	"github.com/canonical/microcloud-cluster-manager/internal/app/management-api/core/auth"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/api/models/v1"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/database/store"
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
			if payload.Description != "" {
				newRemoteCluster.Description = payload.Description
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
		var memberStatuses []models.StatusDistribution
		err := json.Unmarshal(e.MemberStatuses, &memberStatuses)
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
			Description:        e.Description,
			ClusterCertificate: e.ClusterCertificate,
			DiskThreshold:      e.DiskThreshold,
			MemoryThreshold:    e.MemoryThreshold,
			Status:             e.Status,
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
			UIURL:              e.UIURL,
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

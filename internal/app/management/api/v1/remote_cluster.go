package v1

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/canonical/lxd-cluster-manager/internal/pkg/api/models"
	"github.com/canonical/lxd-cluster-manager/internal/pkg/database"
	"github.com/canonical/lxd-cluster-manager/internal/pkg/database/store"
	"github.com/canonical/lxd-cluster-manager/internal/pkg/types"
	"github.com/canonical/lxd/lxd/response"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

var RemoteCluster = types.RouteGroup{
	Prefix: "remote-cluster",
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

func remoteClustersGet(db *database.DB, logger *zap.SugaredLogger) types.EndpointHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		var dbRemoteClusterDetails []store.RemoteClusterWithDetail

		err := db.Transaction(r.Context(), func(ctx context.Context, tx *sqlx.Tx) error {
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

func remoteClusterGet(db *database.DB, logger *zap.SugaredLogger) types.EndpointHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		remoteClusterName, err := url.PathUnescape(mux.Vars(r)["remoteClusterName"])
		if err != nil {
			return response.SmartError(err).Render(w, r)
		}

		var dbRemoteClusterDetails []store.RemoteClusterWithDetail
		err = db.Transaction(r.Context(), func(ctx context.Context, tx *sqlx.Tx) error {
			var err error
			dbRemoteClusterDetails, err = store.GetRemoteClusterWithDetailByName(ctx, tx, remoteClusterName)
			return err
		})

		if err != nil {
			return response.SmartError(err).Render(w, r)
		}

		if len(dbRemoteClusterDetails) == 0 {
			return response.NotFound(fmt.Errorf("RemoteCluster not found")).Render(w, r)
		}

		result, err := toRemoteClustersAPI(dbRemoteClusterDetails)
		if err != nil {
			return response.InternalError(err).Render(w, r)
		}

		return response.SyncResponse(true, result[0]).Render(w, r)
	}
}

func remoteClusterDelete(db *database.DB, logger *zap.SugaredLogger) types.EndpointHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		remoteClusterName, err := url.PathUnescape(mux.Vars(r)["remoteClusterName"])
		if err != nil {
			return response.SmartError(err).Render(w, r)
		}

		err = db.Transaction(r.Context(), func(ctx context.Context, tx *sqlx.Tx) error {
			return store.DeleteRemoteCluster(ctx, tx, remoteClusterName)
		})

		if err != nil {
			return response.SmartError(err).Render(w, r)
		}

		return response.EmptySyncResponse.Render(w, r)
	}
}

func remoteClusterPatch(db *database.DB, logger *zap.SugaredLogger) types.EndpointHandler {
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
			if payload.Status != models.PENDING_APPROVAL && payload.Status != models.ACTIVE {
				return response.BadRequest(fmt.Errorf("invalid status")).Render(w, r)
			}
		}

		err = db.Transaction(r.Context(), func(ctx context.Context, tx *sqlx.Tx) error {
			existingRemoteCluster, err := store.GetRemoteCluster(ctx, tx, remoteClusterName)
			if err != nil {
				return err
			}

			newRemoteCluster := existingRemoteCluster
			if payload.Status != "" {
				if existingRemoteCluster.Status == string(models.PENDING_APPROVAL) && payload.Status == models.ACTIVE {
					newRemoteCluster.JoinedAt = time.Now()
				}
				newRemoteCluster.Status = string(payload.Status)
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
		var ms []models.StatusDistribution
		err := json.Unmarshal(e.MemberStatuses, &ms)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal member statuses: %w", err)
		}

		var is []models.StatusDistribution
		err = json.Unmarshal(e.InstanceStatuses, &is)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal instance statuses: %w", err)
		}

		remoteClusters = append(remoteClusters, models.RemoteCluster{
			Name:               e.Name,
			ClusterCertificate: e.ClusterCertificate,
			Status:             e.Status,
			CPUTotalCount:      e.CPUTotalCount,
			CPULoad1:           e.CPULoad1,
			CPULoad5:           e.CPULoad5,
			CPULoad15:          e.CPULoad15,
			MemoryTotalAmount:  e.MemoryTotalAmount,
			MemoryUsage:        e.MemoryUsage,
			DiskTotalSize:      e.DiskTotalSize,
			DiskUsage:          e.DiskUsage,
			MemberCount:        e.MemberCount,
			MemberStatuses:     ms,
			InstanceCount:      e.InstanceCount,
			InstanceStatuses:   is,
			JoinedAt:           e.ClusterJoinedAt,
			CreatedAt:          e.ClusterCreatedAt,
			LastStatusUpdateAt: e.ClusterUpdatedAt,
		})
	}

	return remoteClusters, nil
}

package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/canonical/lxd/lxd/response"
	"github.com/canonical/microcluster/rest"
	microState "github.com/canonical/microcluster/state"
	"github.com/gorilla/mux"

	"github.com/canonical/lxd-site-manager/internal/api/types"
	"github.com/canonical/lxd-site-manager/internal/database"
	"github.com/canonical/lxd-site-manager/internal/state"
)

func remoteClustersCmd(s *state.ClusterManagerState) rest.Endpoint {
	return rest.Endpoint{
		Path: "remote-clusters",
		Get: rest.EndpointAction{
			Handler:        remoteClustersGet,
			AllowUntrusted: true,
			AccessHandler:  authHandler(s),
		},
	}
}

func remoteClusterCmd(s *state.ClusterManagerState) rest.Endpoint {
	return rest.Endpoint{
		Path: "remote-clusters/{remoteClusterName}",
		Get: rest.EndpointAction{
			Handler:        remoteClusterGet,
			AllowUntrusted: true,
			AccessHandler:  authHandler(s),
		},
		Delete: rest.EndpointAction{
			Handler:        remoteClusterDelete,
			AllowUntrusted: true,
			AccessHandler:  authHandler(s),
		},
		Patch: rest.EndpointAction{
			Handler:        remoteClusterPatch,
			AllowUntrusted: true,
			AccessHandler:  authHandler(s),
		},
	}
}

func remoteClustersGet(s microState.State, r *http.Request) response.Response {
	var dbRemoteClusterDetails []database.CoreRemoteClusterWithDetail

	err := s.Database().Transaction(r.Context(), func(ctx context.Context, tx *sql.Tx) error {
		var err error
		dbRemoteClusterDetails, err = database.GetCoreRemoteClustersWithDetails(ctx, tx)
		return err
	})

	if err != nil {
		return response.SmartError(err)
	}

	result, err := toRemoteClustersAPI(dbRemoteClusterDetails)
	if err != nil {
		return response.InternalError(err)
	}

	return response.SyncResponse(true, result)
}

func remoteClusterGet(s microState.State, r *http.Request) response.Response {
	remoteClusterName, err := url.PathUnescape(mux.Vars(r)["remoteClusterName"])
	if err != nil {
		return response.SmartError(err)
	}

	var dbRemoteClusterDetails []database.CoreRemoteClusterWithDetail
	err = s.Database().Transaction(r.Context(), func(ctx context.Context, tx *sql.Tx) error {
		var err error
		dbRemoteClusterDetails, err = database.GetCoreRemoteClusterWithDetailByName(ctx, tx, remoteClusterName)
		return err
	})

	if err != nil {
		return response.SmartError(err)
	}

	if len(dbRemoteClusterDetails) == 0 {
		return response.NotFound(fmt.Errorf("RemoteCluster not found"))
	}

	result, err := toRemoteClustersAPI(dbRemoteClusterDetails)
	if err != nil {
		return response.InternalError(err)
	}

	return response.SyncResponse(true, result[0])
}

func remoteClusterDelete(s microState.State, r *http.Request) response.Response {
	remoteClusterName, err := url.PathUnescape(mux.Vars(r)["remoteClusterName"])
	if err != nil {
		return response.SmartError(err)
	}

	err = s.Database().Transaction(r.Context(), func(ctx context.Context, tx *sql.Tx) error {
		return database.DeleteCoreRemoteCluster(ctx, tx, remoteClusterName)
	})

	if err != nil {
		return response.SmartError(err)
	}

	return response.EmptySyncResponse
}

func remoteClusterPatch(s microState.State, r *http.Request) response.Response {
	remoteClusterName, err := url.PathUnescape(mux.Vars(r)["remoteClusterName"])
	if err != nil {
		return response.SmartError(err)
	}

	var payload types.RemoteClusterPatch
	err = json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		return response.BadRequest(err)
	}

	if payload.Status != "" {
		if payload.Status != types.PENDING_APPROVAL && payload.Status != types.ACTIVE {
			return response.BadRequest(fmt.Errorf("Invalid status"))
		}
	}

	err = s.Database().Transaction(r.Context(), func(ctx context.Context, tx *sql.Tx) error {
		remoteClusterID, err := database.GetCoreRemoteClusterID(ctx, tx, remoteClusterName)
		if err != nil {
			return err
		}

		existingRemoteClusterDetail, err := database.GetRemoteClusterDetail(ctx, tx, remoteClusterID)
		if err != nil {
			return err
		}

		newRemoteClusterDetail := existingRemoteClusterDetail
		if payload.Status != "" {
			if existingRemoteClusterDetail.Status == string(types.PENDING_APPROVAL) && payload.Status == types.ACTIVE {
				newRemoteClusterDetail.JoinedAt = time.Now()
			}
			newRemoteClusterDetail.Status = string(payload.Status)
		}

		newRemoteClusterDetail.UpdatedAt = time.Now()

		return database.UpdateRemoteClusterDetail(ctx, tx, remoteClusterID, *newRemoteClusterDetail)
	})

	if err != nil {
		return response.SmartError(err)
	}

	return response.EmptySyncResponse
}

func toRemoteClustersAPI(dbEntries []database.CoreRemoteClusterWithDetail) ([]types.RemoteCluster, error) {
	// generate lookup for remoteCluster details
	var remoteClusters []types.RemoteCluster
	for _, e := range dbEntries {
		var ms []types.StatusDistribution
		err := json.Unmarshal([]byte(e.MemberStatuses), &ms)
		if err != nil {
			return nil, fmt.Errorf("Failed to unmarshal member statuses: %w", err)
		}

		var is []types.StatusDistribution
		err = json.Unmarshal([]byte(e.InstanceStatuses), &is)
		if err != nil {
			return nil, fmt.Errorf("Failed to unmarshal instance statuses: %w", err)
		}

		remoteClusters = append(remoteClusters, types.RemoteCluster{
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

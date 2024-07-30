package api

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/canonical/lxd/lxd/response"
	"github.com/canonical/lxd/shared/logger"
	"github.com/canonical/microcluster/rest"
	microTypes "github.com/canonical/microcluster/rest/types"
	"github.com/canonical/microcluster/state"

	"github.com/canonical/lxd-cluster-manager/internal/api/types"
	"github.com/canonical/lxd-cluster-manager/internal/database"
)

var remoteClustersControlCmd = rest.Endpoint{
	Path: "remote-clusters",
	Post: rest.EndpointAction{Handler: remoteClustersPost, AllowUntrusted: true},
}

var remoteClustersStatusCmd = rest.Endpoint{
	Path: "remote-clusters/status",
	Post: rest.EndpointAction{Handler: remoteClustersStatusPost, AllowUntrusted: true},
}

func remoteClustersStatusPost(s state.State, r *http.Request) response.Response {
	if r.TLS == nil || len(r.TLS.PeerCertificates) == 0 {
		logger.Warn("tls is required")
		return response.BadRequest(fmt.Errorf("tls is required"))
	}

	if len(r.TLS.PeerCertificates) != 1 {
		logger.Warn("Expected exactly one peer certificate")
		return response.BadRequest(fmt.Errorf("expected exactly one peer certificate"))
	}
	peerCert := r.TLS.PeerCertificates[0]

	var remoteClusterID int64
	err := s.Database().Transaction(r.Context(), func(ctx context.Context, tx *sql.Tx) error {
		dbRemoteClusters, err := database.GetCoreRemoteClustersWithDetails(ctx, tx)
		if err != nil {
			return err
		}

		for _, dbRemoteCluster := range dbRemoteClusters {
			if dbRemoteCluster.Status == string(types.PENDING_APPROVAL) {
				continue
			}

			dbRemoteClusterCert, err := microTypes.ParseX509Certificate(dbRemoteCluster.ClusterCertificate)
			if err != nil {
				logger.Warn("Failed to parse remoteCluster certificate", logger.Ctx{"remoteCluster": dbRemoteCluster.Name, "err": err})
				continue
			}

			if dbRemoteClusterCert.Certificate.NotAfter.Before(time.Now()) {
				logger.Warn("RemoteCluster certificate is expired", logger.Ctx{"remoteCluster": dbRemoteCluster.Name})
				continue
			}

			// check if public key of dbRemoteCluster matches the peer certificate from the request
			if bytes.Equal(dbRemoteClusterCert.Raw, peerCert.Raw) {
				remoteClusterID = dbRemoteCluster.ID
				break
			}
		}

		return nil
	})

	if err != nil {
		logger.Warn("Failed to get remoteCluster ID", logger.Ctx{"err": err})
		return response.SmartError(err)
	}

	if remoteClusterID == 0 {
		logger.Warn("cluster not found")
		return response.NotFound(fmt.Errorf("cluster not found"))
	}

	payload := types.RemoteClusterStatusPost{}
	err = json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		return response.BadRequest(err)
	}

	err = s.Database().Transaction(r.Context(), func(ctx context.Context, tx *sql.Tx) error {
		dbRemoteCluster, err := database.GetRemoteClusterDetail(ctx, tx, remoteClusterID)
		if err != nil {
			return err
		}

		dbRemoteCluster.CPULoad1 = payload.CPULoad1
		dbRemoteCluster.CPULoad5 = payload.CPULoad5
		dbRemoteCluster.CPULoad15 = payload.CPULoad15
		dbRemoteCluster.CPUTotalCount = payload.CPUTotalCount
		dbRemoteCluster.DiskTotalSize = payload.DiskTotalSize
		dbRemoteCluster.DiskUsage = payload.DiskUsage
		dbRemoteCluster.InstanceCount, dbRemoteCluster.InstanceStatuses = parseStatusDistribution(payload.InstanceStatuses)
		dbRemoteCluster.MemberCount, dbRemoteCluster.MemberStatuses = parseStatusDistribution(payload.MemberStatuses)
		dbRemoteCluster.MemoryTotalAmount = payload.MemoryTotalAmount
		dbRemoteCluster.MemoryUsage = payload.MemoryUsage
		dbRemoteCluster.UpdatedAt = time.Now()

		err = database.UpdateRemoteClusterDetail(ctx, tx, remoteClusterID, *dbRemoteCluster)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		logger.Warn("Failed to update remoteCluster status", logger.Ctx{"remoteCluster": remoteClusterID, "err": err})
		return response.SmartError(err)
	}

	memberAddresses, err := getClusterManagerAddresses(r.Context(), s)
	if err != nil {
		return response.InternalError(err)
	}

	return response.SyncResponse(true, types.RemoteClusterStatusPostResponse{
		ClusterManagerAddresses: memberAddresses,
	})
}

func parseStatusDistribution(statuses []types.StatusDistribution) (int64, string) {
	if len(statuses) == 0 {
		return 0, "[]"
	}

	parsedStatuses, err := json.Marshal(statuses)
	if err != nil {
		return 0, "[]"
	}

	var total int64
	for _, s := range statuses {
		total += s.Count
	}

	return total, string(parsedStatuses)
}

func remoteClustersPost(s state.State, r *http.Request) response.Response {
	payload := types.RemoteClusterPost{}

	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		return response.BadRequest(err)
	}

	if payload.ClusterName == "" {
		return response.BadRequest(fmt.Errorf("remoteCluster name is required"))
	}

	if payload.ClusterCertificate == "" {
		return response.BadRequest(fmt.Errorf("remoteCluster certificate is required"))
	}

	// get token secret for HMAC verification
	var token *database.CoreRemoteClusterToken
	err = s.Database().Transaction(r.Context(), func(ctx context.Context, tx *sql.Tx) error {
		var err error
		token, err = database.GetCoreRemoteClusterToken(ctx, tx, payload.ClusterName)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return response.SmartError(err)
	}

	// check if token has expired
	if time.Now().After(token.Expiry) {
		return response.Forbidden(fmt.Errorf("token has expired"))
	}

	// verify HMAC
	hmacOK, err := verifyHMAC(payload, r, token.Secret)
	if err != nil || !hmacOK {
		return response.Forbidden(err)
	}

	// Create remoteCluster entry and delete token in a single db transaction
	err = s.Database().Transaction(r.Context(), func(ctx context.Context, tx *sql.Tx) error {
		// create remoteCluster entry
		remoteClusterID, err := database.CreateCoreRemoteCluster(ctx, tx, database.CoreRemoteCluster{
			Name:               payload.ClusterName,
			ClusterCertificate: payload.ClusterCertificate,
		})

		if err != nil {
			return err
		}

		// create relevant remoteCluster details
		_, err = database.CreateRemoteClusterDetail(ctx, tx, database.RemoteClusterDetail{
			Status:              string(types.PENDING_APPROVAL),
			CoreRemoteClusterID: remoteClusterID,
			JoinedAt:            time.Now(),
			MemberStatuses:      "[]",
			InstanceStatuses:    "[]",
		})

		if err != nil {
			return err
		}

		// delete remoteCluster token
		return database.DeleteCoreRemoteClusterToken(ctx, tx, payload.ClusterName)
	})

	if err != nil {
		return response.SmartError(err)
	}

	return response.EmptySyncResponse
}

func verifyHMAC(payload types.RemoteClusterPost, r *http.Request, secret string) (bool, error) {
	reqBody, err := json.Marshal(payload)
	if err != nil {
		return false, fmt.Errorf("failed to marshal payload: %v", err)
	}

	sig := r.Header.Get("X-CLUSTER-SIGNATURE")
	if sig == "" {
		return false, fmt.Errorf("missing signature header")
	}

	decodedSig, err := base64.StdEncoding.DecodeString(sig)
	if err != nil {
		return false, fmt.Errorf("failed to decode signature: %v", err)
	}

	// recompute the HMAC
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(reqBody)
	expectedMac := mac.Sum(nil)

	return hmac.Equal(decodedSig, expectedMac), nil
}

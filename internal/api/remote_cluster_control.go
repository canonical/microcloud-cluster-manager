package api

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/canonical/lxd/lxd/request"
	"github.com/canonical/lxd/lxd/response"
	"github.com/canonical/lxd/shared/logger"
	"github.com/canonical/microcluster/v2/rest"
	microTypes "github.com/canonical/microcluster/v2/rest/types"
	microState "github.com/canonical/microcluster/v2/state"

	"github.com/canonical/lxd-cluster-manager/internal/api/types"
	"github.com/canonical/lxd-cluster-manager/internal/database"
	"github.com/canonical/lxd-cluster-manager/internal/state"
)

func remoteClustersControlCmd(s *state.ClusterManagerState) rest.Endpoint {
	return rest.Endpoint{
		Path: "remote-clusters",
		Post: rest.EndpointAction{
			Handler:        remoteClustersPost(s),
			AllowUntrusted: true,
		},
	}
}

func remoteClustersStatusCmd(s *state.ClusterManagerState) rest.Endpoint {
	return rest.Endpoint{
		Path: "remote-clusters/status",
		Post: rest.EndpointAction{
			Handler:        remoteClustersStatusPost(s),
			AllowUntrusted: true,
			AccessHandler:  mtlsAuthHandler(s),
		},
	}
}

func remoteClustersStatusPost(managerState *state.ClusterManagerState) types.EndpointHandler {
	return func(clusterState microState.State, r *http.Request) response.Response {
		payload := types.RemoteClusterStatusPost{}
		err := json.NewDecoder(r.Body).Decode(&payload)
		if err != nil {
			return response.BadRequest(err)
		}

		remoteClusterID, err := request.GetCtxValue[int64](r.Context(), ctxRemoteClusterID)
		if err != nil {
			return response.SmartError(err)
		}

		err = managerState.Database.Transaction(r.Context(), func(ctx context.Context, tx *sql.Tx) error {
			dbRemoteCluster, err := database.GetRemoteClusterDetail(ctx, tx, remoteClusterID)
			if err != nil {
				return err
			}

			if dbRemoteCluster.Status == string(types.PENDING_APPROVAL) {
				return fmt.Errorf("remote cluster is pending approval")
			}

			dbRemoteCluster.Put(payload)

			err = database.UpdateRemoteClusterDetail(ctx, tx, remoteClusterID, *dbRemoteCluster)
			if err != nil {
				return err
			}

			return nil
		})

		if err != nil {
			logger.Warn("Failed to update remote cluster status", logger.Ctx{"remote cluster": remoteClusterID, "err": err})
			return response.SmartError(err)
		}

		memberAddresses, err := getClusterManagerAddresses(r.Context(), clusterState)
		if err != nil {
			return response.InternalError(err)
		}

		return response.SyncResponse(true, types.RemoteClusterStatusPostResponse{
			ClusterManagerAddresses: memberAddresses,
		})
	}
}

func remoteClustersPost(managerState *state.ClusterManagerState) types.EndpointHandler {
	return func(clusterState microState.State, r *http.Request) response.Response {
		payload := types.RemoteClusterPost{}

		err := json.NewDecoder(r.Body).Decode(&payload)
		if err != nil {
			return response.BadRequest(err)
		}

		if payload.ClusterName == "" {
			return response.BadRequest(fmt.Errorf("remote cluster name is required"))
		}

		if payload.ClusterCertificate == "" {
			return response.BadRequest(fmt.Errorf("remote cluster certificate is required"))
		}

		cert, err := microTypes.ParseX509Certificate(payload.ClusterCertificate)
		if err != nil {
			return response.BadRequest(fmt.Errorf("invalid certificate: %v", err))
		}

		// get token secret for HMAC verification
		var token *database.CoreRemoteClusterToken
		err = managerState.Database.Transaction(r.Context(), func(ctx context.Context, tx *sql.Tx) error {
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

		// Create remote cluster entry and delete token in a single db transaction
		var remoteClusterID int64
		err = managerState.Database.Transaction(r.Context(), func(ctx context.Context, tx *sql.Tx) error {
			var err error
			// create remote cluster entry
			remoteClusterID, err = database.CreateCoreRemoteCluster(ctx, tx, database.CoreRemoteCluster{
				Name:               payload.ClusterName,
				ClusterCertificate: payload.ClusterCertificate,
			})

			if err != nil {
				return err
			}

			// create relevant remote cluster details
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

			// delete remote cluster token
			return database.DeleteCoreRemoteClusterToken(ctx, tx, payload.ClusterName)
		})

		if err != nil {
			return response.SmartError(err)
		}

		err = managerState.CertificatesCache.AddCertificate(cert.Certificate, remoteClusterID)
		if err != nil {
			return response.InternalError(err)
		}

		return response.EmptySyncResponse
	}
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

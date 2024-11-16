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
			Method:  http.MethodPost,
			Handler: remoteClustersPost,
		},
		{
			Path:    "status",
			Method:  http.MethodPost,
			Handler: remoteClusterStatusPost,
		},
		{
			Path:    "{remoteClusterName}",
			Method:  http.MethodDelete,
			Handler: remoteClusterDelete,
		},
	},
}

func remoteClustersPost(db *database.DB, logger *zap.SugaredLogger) types.EndpointHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		payload := models.RemoteClusterPost{}

		err := json.NewDecoder(r.Body).Decode(&payload)
		if err != nil {
			return response.BadRequest(err).Render(w, r)
		}

		if payload.ClusterName == "" {
			return response.BadRequest(fmt.Errorf("remote cluster name is required")).Render(w, r)
		}

		if payload.ClusterCertificate == "" {
			return response.BadRequest(fmt.Errorf("remote cluster certificate is required")).Render(w, r)
		}

		// TODO: parse certificate
		// cert, err := microTypes.ParseX509Certificate(payload.ClusterCertificate)
		// if err != nil {
		// 	return response.BadRequest(fmt.Errorf("invalid certificate: %v", err)).Render(w, r)
		// }

		// get token secret for HMAC verification
		var token *store.RemoteClusterToken
		err = db.Transaction(r.Context(), func(ctx context.Context, tx *sqlx.Tx) error {
			var err error
			token, err = store.GetRemoteClusterToken(ctx, tx, payload.ClusterName)
			if err != nil {
				return err
			}

			return nil
		})

		if err != nil {
			return response.SmartError(err).Render(w, r)
		}

		// check if token has expired
		if time.Now().After(token.Expiry) {
			return response.Forbidden(fmt.Errorf("token has expired")).Render(w, r)
		}

		// TODO: verify HMAC
		// hmacOK, err := verifyHMAC(payload, r, token.Secret)
		// if err != nil || !hmacOK {
		// 	return response.Forbidden(err)
		// }

		// Create remote cluster entry and delete token in a single db transaction
		err = db.Transaction(r.Context(), func(ctx context.Context, tx *sqlx.Tx) error {
			// create remote cluster entry
			newRemoteCluster, err := store.CreateRemoteCluster(ctx, tx, store.RemoteCluster{
				Name:               payload.ClusterName,
				ClusterCertificate: payload.ClusterCertificate,
				Status:             string(models.PENDING_APPROVAL),
			})

			if err != nil {
				return err
			}

			// create relevant remote cluster details
			_, err = store.CreateRemoteClusterDetail(ctx, tx, store.RemoteClusterDetail{
				RemoteClusterID:  newRemoteCluster.ID,
				MemberStatuses:   json.RawMessage("[]"),
				InstanceStatuses: json.RawMessage("[]"),
			})

			if err != nil {
				return err
			}

			// delete remote cluster token
			return store.DeleteRemoteClusterToken(ctx, tx, payload.ClusterName)
		})

		if err != nil {
			return response.SmartError(err).Render(w, r)
		}

		if err != nil {
			return response.InternalError(err).Render(w, r)
		}

		return response.EmptySyncResponse.Render(w, r)
	}
}

// apply mtls for this endpoint
func remoteClusterStatusPost(db *database.DB, logger *zap.SugaredLogger) types.EndpointHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		payload := models.RemoteClusterStatusPost{}
		err := json.NewDecoder(r.Body).Decode(&payload)
		if err != nil {
			return response.BadRequest(err).Render(w, r)
		}

		// TODO: mtls verification
		// remoteClusterID, err := request.GetCtxValue[int64](r.Context(), ctxRemoteClusterID)
		// if err != nil {
		// 	return response.SmartError(err).Render(w, r)
		// }

		var remoteClusterID int
		err = db.Transaction(r.Context(), func(ctx context.Context, tx *sqlx.Tx) error {
			// TODO: get remote cluster by certificate after mtls logic is in place
			dbRemoteCluster, err := store.GetRemoteCluster(ctx, tx, payload.ClusterName)
			if err != nil {
				return err
			}

			remoteClusterID = dbRemoteCluster.ID
			dbRemoteClusterDetail, err := store.GetRemoteClusterDetail(ctx, tx, dbRemoteCluster.ID)
			if err != nil {
				return err
			}

			if dbRemoteCluster.Status == string(models.PENDING_APPROVAL) {
				return fmt.Errorf("remote cluster is pending approval")
			}

			dbRemoteClusterDetail.Put(payload)

			err = store.UpdateRemoteClusterDetail(ctx, tx, dbRemoteCluster.ID, *dbRemoteClusterDetail)
			if err != nil {
				return err
			}

			return nil
		})

		if err != nil {
			logger.Warnw("Failed to update remote cluster status", "remote cluster", remoteClusterID, "err", err)
			return response.SmartError(err).Render(w, r)
		}

		// TODO: determine next update time

		return response.SyncResponse(true, models.RemoteClusterStatusPostResponse{
			NextUpdateInSeconds: time.Now().Local().String(),
		}).Render(w, r)
	}
}

// TODO: use cert fingerprint instead of path param for deleting cluster
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

package v1

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/canonical/lxd/lxd/response"
	"github.com/canonical/lxd/shared"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"

	"github.com/canonical/lxd-cluster-manager/internal/pkg/api/models"
	"github.com/canonical/lxd-cluster-manager/internal/pkg/database"
	"github.com/canonical/lxd-cluster-manager/internal/pkg/database/store"
	"github.com/canonical/lxd-cluster-manager/internal/pkg/types"
)

var RemoteClusterJoinToken = types.RouteGroup{
	Prefix: "remote-cluster-join-token",
	Endpoints: []types.Endpoint{
		{
			Method:  http.MethodPost,
			Handler: tokenPost,
		},
		{
			Method:  http.MethodGet,
			Handler: tokensGet,
		},
		{
			Path:    "{remoteClusterName}",
			Method:  http.MethodDelete,
			Handler: tokenDelete,
		},
	},
}

func tokenPost(db *database.DB, _ *zap.SugaredLogger) types.EndpointHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		payload := models.RemoteClusterTokenPost{}

		err := json.NewDecoder(r.Body).Decode(&payload)
		if err != nil {
			return response.BadRequest(err).Render(w, r)
		}

		// default expiry to 1 day if not set
		if payload.Expiry == (time.Time{}) {
			payload.Expiry = time.Now().Add(time.Hour * 24)
		}

		if payload.Expiry.Before(time.Now()) {
			return response.BadRequest(fmt.Errorf("expiry date must be in the future")).Render(w, r)
		}

		if payload.ClusterName == "" {
			return response.BadRequest(fmt.Errorf("cluster name is required")).Render(w, r)
		}

		secret, err := shared.RandomCryptoString()
		if err != nil {
			return response.InternalError(err).Render(w, r)
		}

		// store token details in the database
		err = db.Transaction(r.Context(), func(ctx context.Context, tx *sqlx.Tx) error {
			var err error
			tokenData := store.RemoteClusterToken{
				ClusterName: payload.ClusterName,
				Secret:      secret,
				Expiry:      payload.Expiry,
				CreatedAt:   time.Now(),
			}
			_, err = store.CreateRemoteClusterToken(ctx, tx, tokenData)

			return err
		})

		if err != nil {
			return response.SmartError(err).Render(w, r)
		}

		// create the token to be sent to LXD

		// TODO: get the cluster cert
		// if err != nil {
		// 	return response.InternalError(err).Render(w, r)
		// }

		// TODO: get the control service address for the token payload
		// memberAddresses, err := getClusterManagerAddresses(r.Context(), s)
		// if err != nil {
		// 	return response.InternalError(err)
		// }

		token := models.RemoteClusterTokenBody{
			Secret:      secret,
			ExpiresAt:   payload.Expiry,
			Address:     "localhost",
			ServerName:  payload.ClusterName,
			Fingerprint: "abc",
		}

		encodedToken, err := token.Encode()
		if err != nil {
			return response.InternalError(err).Render(w, r)
		}

		return response.SyncResponse(true, models.RemoteClusterTokenPostResponse{Token: encodedToken}).Render(w, r)
	}
}

func tokensGet(db *database.DB, _ *zap.SugaredLogger) types.EndpointHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		var tokens []store.RemoteClusterToken
		err := db.Transaction(r.Context(), func(ctx context.Context, tx *sqlx.Tx) error {
			var err error
			tokens, err = store.GetRemoteClusterTokens(ctx, tx)
			return err
		})

		if err != nil {
			return response.SmartError(err).Render(w, r)
		}

		var responseTokens []models.RemoteClusterToken
		for _, token := range tokens {
			responseTokens = append(responseTokens, models.RemoteClusterToken{
				Expiry:      token.Expiry,
				ClusterName: token.ClusterName,
				CreateAt:    token.CreatedAt,
			})
		}

		return response.SyncResponse(true, responseTokens).Render(w, r)
	}
}

func tokenDelete(db *database.DB, _ *zap.SugaredLogger) types.EndpointHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		remoteClusterName, err := url.PathUnescape(mux.Vars(r)["remoteClusterName"])
		if err != nil {
			return response.SmartError(err).Render(w, r)
		}

		if remoteClusterName == "" {
			return response.BadRequest(fmt.Errorf("cluster name is required")).Render(w, r)
		}

		err = db.Transaction(r.Context(), func(ctx context.Context, tx *sqlx.Tx) error {
			return store.DeleteRemoteClusterToken(ctx, tx, remoteClusterName)
		})

		if err != nil {
			return response.SmartError(err).Render(w, r)
		}

		return response.SyncResponse(true, nil).Render(w, r)
	}
}

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
	"github.com/canonical/lxd/shared"
	"github.com/canonical/microcloud-cluster-manager/internal/app/management-api/core/auth"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/api/models/v1"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/database/store"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/types"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
)

// RemoteClusterJoinToken is the remote cluster join token endpoint group.
var RemoteClusterJoinToken = types.RouteGroup{
	Prefix: "remote-cluster-join-token",
	Middlewares: []types.RouteMiddleware{
		auth.AuthMiddleware,
	},
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

func tokenPost(rc types.RouteConfig) types.EndpointHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		payload := models.RemoteClusterTokenPost{}

		err := json.NewDecoder(r.Body).Decode(&payload)
		if err != nil {
			return response.BadRequest(err).Render(w, r)
		}

		// default expiry to 1 day if not set
		if time.Time.Equal(payload.Expiry, time.Time{}) {
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

		// create the token to be sent to LXD
		cert, err := rc.Env.ClusterConnectorCert.PublicKeyX509()
		if err != nil {
			return response.InternalError(err).Render(w, r)
		}

		// get the cluster-connector service address for the token payload
		clusterConnectorAddress := rc.Env.ClusterConnectorAddress

		token := models.RemoteClusterTokenBody{
			Secret:      secret,
			ExpiresAt:   payload.Expiry,
			Addresses:   []string{clusterConnectorAddress},
			ServerName:  payload.ClusterName,
			Fingerprint: shared.CertFingerprint(cert),
		}
		encodedToken, err := token.Encode()
		if err != nil {
			return response.InternalError(err).Render(w, r)
		}

		// store token details in the database
		err = rc.DB.Transaction(r.Context(), func(ctx context.Context, tx *sqlx.Tx) error {
			var err error
			isNameTaken, err := store.RemoteClusterExists(ctx, tx, payload.ClusterName)
			if err != nil {
				return err
			}
			if isNameTaken {
				return fmt.Errorf("cluster name already exists")
			}

			tokenData := store.RemoteClusterToken{
				ClusterName:  payload.ClusterName,
				Description:  payload.Description,
				EncodedToken: encodedToken,
				Expiry:       payload.Expiry,
				CreatedAt:    time.Now(),
			}
			_, err = store.CreateRemoteClusterToken(ctx, tx, tokenData)

			return err
		})

		if err != nil {
			return response.SmartError(err).Render(w, r)
		}

		if err != nil {
			return response.InternalError(err).Render(w, r)
		}

		return response.SyncResponse(true, models.RemoteClusterTokenPostResponse{Token: encodedToken}).Render(w, r)
	}
}

func tokensGet(rc types.RouteConfig) types.EndpointHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		var tokens []store.RemoteClusterToken
		err := rc.DB.Transaction(r.Context(), func(ctx context.Context, tx *sqlx.Tx) error {
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
				Description: token.Description,
				CreateAt:    token.CreatedAt,
			})
		}

		sort.Slice(responseTokens, func(i, j int) bool {
			return responseTokens[i].CreateAt.Before(responseTokens[j].CreateAt)
		})

		return response.SyncResponse(true, responseTokens).Render(w, r)
	}
}

func tokenDelete(rc types.RouteConfig) types.EndpointHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		remoteClusterName, err := url.PathUnescape(mux.Vars(r)["remoteClusterName"])
		if err != nil {
			return response.SmartError(err).Render(w, r)
		}

		if remoteClusterName == "" {
			return response.BadRequest(fmt.Errorf("cluster name is required")).Render(w, r)
		}

		err = rc.DB.Transaction(r.Context(), func(ctx context.Context, tx *sqlx.Tx) error {
			return store.DeleteRemoteClusterToken(ctx, tx, remoteClusterName)
		})

		if err != nil {
			return response.SmartError(err).Render(w, r)
		}

		return response.SyncResponse(true, nil).Render(w, r)
	}
}

package v1

import (
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
	"time"

	"github.com/canonical/lxd/lxd/response"
	"github.com/canonical/lxd/shared"
	"github.com/canonical/microcloud-cluster-manager/internal/app/management-api/core/auth"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/api/models/v1"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/database/store"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/logger"
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
			Path:    "{remoteClusterName}/tunnel/{path:.*}",
			Method:  http.MethodGet,
			Handler: remoteClusterTunnel,
		},
		{
			Path:    "{remoteClusterName}/tunnel/{path:.*}",
			Method:  http.MethodPost,
			Handler: remoteClusterTunnel,
		},
		{
			Path:    "{remoteClusterName}/tunnel/{path:.*}",
			Method:  http.MethodPut,
			Handler: remoteClusterTunnel,
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
			newRemoteCluster.Description = payload.Description
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
			UIURL:              e.UIURL,
			TunnelRegistered:   e.TunnelManagerMemberURL != "",
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

func remoteClusterTunnel(rc types.RouteConfig) types.EndpointHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		remoteClusterName, err := url.PathUnescape(mux.Vars(r)["remoteClusterName"])
		if err != nil {
			return response.SmartError(err).Render(w, r)
		}

		var clusterDetails *store.RemoteClusterDetail
		err = rc.DB.Transaction(r.Context(), func(ctx context.Context, tx *sqlx.Tx) error {
			cluster, err := store.GetRemoteCluster(ctx, tx, remoteClusterName)
			if err != nil {
				return err
			}

			clusterDetails, err = store.GetRemoteClusterDetail(ctx, tx, cluster.ID)
			return err
		})
		if err != nil {
			return response.SmartError(err).Render(w, r)
		}

		if clusterDetails.TunnelManagerMemberURL == "" {
			return response.BadRequest(fmt.Errorf("remote cluster does not have an active tunnel")).Render(w, r)
		}

		reqURL := clusterDetails.TunnelManagerMemberURL + r.URL.Path
		if r.URL.RawQuery != "" {
			reqURL += "?" + r.URL.RawQuery
		}

		sessionProvider, ok := r.Context().Value(auth.SessionCredentialsProviderKey).(*auth.SessionCredentialsProvider)
		if !ok {
			return response.InternalError(fmt.Errorf("failed to get session credentials provider from the request context")).Render(w, r)
		}

		accessToken, err := sessionProvider.GetAccessToken()
		if err != nil {
			return response.InternalError(fmt.Errorf("failed to get access token: %w", err)).Render(w, r)
		}

		userSecret, err := sessionProvider.GetUserSecret()
		if err != nil {
			return response.InternalError(fmt.Errorf("failed to get user secret: %w", err)).Render(w, r)
		}

		resp, err := sendTunnelRequest(r, accessToken, userSecret, reqURL, rc)
		if err != nil {
			return response.SmartError(err).Render(w, r)
		}

		needsTokenRefresh := resp.StatusCode == 401 || resp.StatusCode == 403
		if needsTokenRefresh {
			refreshCtx, cancelRefresh := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancelRefresh()
			accessToken, err = sessionProvider.RefreshSessionTokens(refreshCtx)
			if err != nil {
				return response.InternalError(fmt.Errorf("failed to refresh access token: %w", err)).Render(w, r)
			}

			resp, err = sendTunnelRequest(r, accessToken, userSecret, reqURL, rc)
			if err != nil {
				return response.InternalError(fmt.Errorf("failed to send request after token refresh: %w", err)).Render(w, r)
			}
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
}

func sendTunnelRequest(r *http.Request, accessToken string, userSecret string, reqURL string, rc types.RouteConfig) (*http.Response, error) {
	req, err := http.NewRequestWithContext(r.Context(), r.Method, reqURL, r.Body)
	if err != nil {
		return nil, err
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

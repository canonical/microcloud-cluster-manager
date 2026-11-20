package v1

import (
	"bytes"
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
	"strings"
	"time"

	"github.com/canonical/lxd/lxd/response"
	"github.com/canonical/lxd/shared"
	"github.com/canonical/lxd/shared/api"
	"github.com/canonical/microcloud-cluster-manager/internal/app/management-api/core/auth"
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
			Path:    "{remoteClusterName}/cluster-links/{clusterLinkName}",
			Method:  http.MethodGet,
			Handler: remoteClusterLinkGet,
		},
		{
			Path:    "{remoteClusterName}/cluster-links/{clusterLinkName}/state",
			Method:  http.MethodGet,
			Handler: remoteClusterLinkStateGet,
		},
		{
			Path:    "{remoteClusterName}/replicators",
			Method:  http.MethodGet,
			Handler: remoteClusterReplicatorsGet,
		},
		{
			Path:    "{remoteClusterName}/replicators",
			Method:  http.MethodPost,
			Handler: remoteClusterReplicatorsPost,
		},
		{
			Path:    "{remoteClusterName}/replicators/{replicatorName}",
			Method:  http.MethodGet,
			Handler: remoteClusterReplicatorGet,
		},
		{
			Path:    "{remoteClusterName}/replicators/{replicatorName}",
			Method:  http.MethodPut,
			Handler: remoteClusterReplicatorPut,
		},
		{
			Path:    "{remoteClusterName}/replicators/{replicatorName}",
			Method:  http.MethodDelete,
			Handler: remoteClusterReplicatorDelete,
		},
		{
			Path:    "{remoteClusterName}/replicators/{replicatorName}/state",
			Method:  http.MethodGet,
			Handler: remoteClusterReplicatorStateGet,
		},
		{
			Path:    "{remoteClusterName}/replicators/{replicatorName}/state",
			Method:  http.MethodPut,
			Handler: remoteClusterReplicatorStatePut,
		},
		{
			Path:    "{remoteClusterName}/projects",
			Method:  http.MethodGet,
			Handler: remoteClusterProjectsGet,
		},
		{
			Path:    "{remoteClusterName}/projects",
			Method:  http.MethodPost,
			Handler: remoteClusterProjectsPost,
		},
		{
			Path:    "{remoteClusterName}/projects/{projectName}",
			Method:  http.MethodGet,
			Handler: remoteClusterProjectGet,
		},
		{
			Path:    "{remoteClusterName}/projects/{projectName}",
			Method:  http.MethodPut,
			Handler: remoteClusterProjectPut,
		},
		{
			Path:    "{remoteClusterName}/projects/{projectName}/state",
			Method:  http.MethodGet,
			Handler: remoteClusterProjectStateGet,
		},
		{
			Path:    "{remoteClusterName}/projects/{projectName}/state",
			Method:  http.MethodPut,
			Handler: remoteClusterProjectStatePut,
		},
		{
			Path:    "{remoteClusterName}/operations/{operationID}",
			Method:  http.MethodGet,
			Handler: remoteClusterOperationGet,
		},
		{
			Path:    "{remoteClusterName}/auth/identities",
			Method:  http.MethodGet,
			Handler: remoteClusterIdentitiesGet,
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
		remoteClusterName, err := request.RemoteClusterName(r)
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
		remoteClusterName, err := request.RemoteClusterName(r)
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
		remoteClusterName, err := request.RemoteClusterName(r)
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

func remoteClusterAuthGroupsGet(rc types.RouteConfig) types.EndpointHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		return forwardToClusterConnector(rc, w, r, "/auth-groups")
	}
}

func remoteClusterIdentitiesGet(rc types.RouteConfig) types.EndpointHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		return forwardToClusterConnector(rc, w, r, "/auth/identities")
	}
}

func remoteClusterLinkGet(rc types.RouteConfig) types.EndpointHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		clusterLinkName, err := request.PathParam(r, request.ClusterLinkNameParam, "cluster link name")
		if err != nil {
			return response.SmartError(err).Render(w, r)
		}

		clusterConnectorPath := "/cluster-links/" + url.PathEscape(clusterLinkName)
		return forwardToClusterConnector(rc, w, r, clusterConnectorPath)
	}
}

func remoteClusterLinkStateGet(rc types.RouteConfig) types.EndpointHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		clusterLinkName, err := url.PathUnescape(mux.Vars(r)["clusterLinkName"])
		if err != nil {
			return response.SmartError(err).Render(w, r)
		}

		clusterConnectorPath := "/cluster-links/" + url.PathEscape(clusterLinkName) + "/state"
		return forwardToClusterConnector(rc, w, r, clusterConnectorPath)
	}
}

// requestClusterMember sends a request without a body for the given remote cluster to the cluster connector member
// holding its tunnel and decodes the metadata of the response into target, unless target is nil. A response larger
// than maxResponseBytes is rejected, unless maxResponseBytes is zero. A non-2xx response is returned as an error
// carrying the upstream status code.
func requestClusterMember(rc types.RouteConfig, r *http.Request, method string, cluster store.RemoteClusterWithDetail, clusterConnectorPath string, target any, maxResponseBytes int64) error {
	resp, err := sendToClusterMember(rc, r, method, cluster.Name, cluster.TunnelManagerMemberURL, clusterConnectorPath, nil)
	if err != nil {
		return err
	}

	defer func() {
		err := resp.Body.Close()
		if err != nil {
			logger.Log.Errorf("failed to close response body: %v", err)
		}
	}()

	var reader io.Reader = resp.Body
	if maxResponseBytes > 0 {
		// Read one byte more than allowed to tell a response that is exactly at the limit from a larger one.
		reader = io.LimitReader(resp.Body, maxResponseBytes+1)
	}

	body, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("Failed to read response for cluster %q: %w", cluster.Name, err)
	}

	if maxResponseBytes > 0 && int64(len(body)) > maxResponseBytes {
		return api.StatusErrorf(http.StatusBadGateway, "Response for cluster %q is larger than %d bytes", cluster.Name, maxResponseBytes)
	}

	var lxdResponse struct {
		Metadata json.RawMessage `json:"metadata"`
		Error    string          `json:"error"`
	}

	decodeErr := json.Unmarshal(body, &lxdResponse)

	if !request.IsSuccessStatus(resp.StatusCode) {
		if decodeErr != nil || lxdResponse.Error == "" {
			return api.StatusErrorf(resp.StatusCode, "Request %s %s for cluster %q failed with HTTP status %d", method, clusterConnectorPath, cluster.Name, resp.StatusCode)
		}

		return api.StatusErrorf(resp.StatusCode, "%s", lxdResponse.Error)
	}

	if target == nil {
		return nil
	}

	if decodeErr != nil {
		return api.StatusErrorf(http.StatusBadGateway, "Invalid response for cluster %q: %v", cluster.Name, decodeErr)
	}

	err = json.Unmarshal(lxdResponse.Metadata, target)
	if err != nil {
		return api.StatusErrorf(http.StatusBadGateway, "Invalid response metadata for cluster %q: %v", cluster.Name, err)
	}

	return nil
}

// forwardToClusterConnector forwards the request to the cluster connector path of the remote cluster named in the
// request path and streams back the response.
func forwardToClusterConnector(rc types.RouteConfig, w http.ResponseWriter, r *http.Request, clusterConnectorPath string) error {
	remoteClusterName, err := request.RemoteClusterName(r)
	if err != nil {
		return response.SmartError(err).Render(w, r)
	}

	return forwardToRemoteClusterConnector(rc, w, r, remoteClusterName, clusterConnectorPath)
}

// forwardToRemoteClusterConnector forwards the request to the cluster connector path of the given remote cluster and
// streams back the response.
func forwardToRemoteClusterConnector(rc types.RouteConfig, w http.ResponseWriter, r *http.Request, remoteClusterName string, clusterConnectorPath string) error {
	// Buffer the body so that it can be sent again if the request needs to be retried after a token refresh.
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return response.SmartError(fmt.Errorf("failed to read request body: %w", err)).Render(w, r)
	}

	defer func() {
		err := r.Body.Close()
		if err != nil {
			logger.Log.Warnw("Failed to close request body", "error", err)
		}
	}()

	memberURL, err := getClusterMemberURL(r.Context(), rc, remoteClusterName)
	if err != nil {
		return response.SmartError(err).Render(w, r)
	}

	resp, err := sendToClusterMember(rc, r, r.Method, remoteClusterName, memberURL, clusterConnectorPath, body)
	if err != nil {
		return response.SmartError(err).Render(w, r)
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

// getClusterMemberURL looks up the Connector member URL stored for the named remote cluster.
func getClusterMemberURL(ctx context.Context, rc types.RouteConfig, remoteClusterName string) (string, error) {
	var memberURL string
	err := rc.DB.Transaction(ctx, func(ctx context.Context, tx *sqlx.Tx) error {
		cluster, err := store.GetRemoteCluster(ctx, tx, remoteClusterName)
		if err != nil {
			return err
		}

		clusterDetails, err := store.GetRemoteClusterDetail(ctx, tx, cluster.ID)
		if err != nil {
			return err
		}

		memberURL = clusterDetails.TunnelManagerMemberURL
		return nil
	})
	if err != nil {
		return "", err
	}

	return memberURL, nil
}

// sendToClusterMember sends a request for the given remote cluster to the cluster connector member holding its
// tunnel, at the given member URL. Requests without side effects are retried once with refreshed session tokens after
// a 401 or 403 response. The caller must close the body of the returned response.
func sendToClusterMember(rc types.RouteConfig, r *http.Request, method string, remoteClusterName string, memberURL string, clusterConnectorPath string, body []byte) (*http.Response, error) {
	if memberURL == "" {
		return nil, api.StatusErrorf(http.StatusBadRequest, "remote cluster %q does not have an active tunnel", remoteClusterName)
	}

	// Top-level connector paths carry their cluster scope in the query.
	if !strings.HasPrefix(clusterConnectorPath, "/1.0/") {
		clusterConnectorPath = "/1.0/remote-cluster/" + url.PathEscape(remoteClusterName) + clusterConnectorPath
	}
	reqURL := memberURL + clusterConnectorPath

	sessionProvider, ok := r.Context().Value(auth.SessionCredentialsProviderKey).(*auth.SessionCredentialsProvider)
	if !ok {
		return nil, api.StatusErrorf(http.StatusInternalServerError, "failed to get session credentials provider from the request context")
	}

	accessToken, err := sessionProvider.GetAccessToken()
	if err != nil {
		return nil, api.StatusErrorf(http.StatusInternalServerError, "failed to get access token: %v", err)
	}

	userSecret, err := sessionProvider.GetUserSecret()
	if err != nil {
		return nil, api.StatusErrorf(http.StatusInternalServerError, "failed to get user secret: %v", err)
	}

	resp, err := sendInternalRequest(r, method, accessToken, userSecret, reqURL, rc, body)
	if err != nil {
		return nil, err
	}

	// Only replay requests that have no side effects. A 401/403 from a non-idempotent request may be the
	// result of a partially applied change, so replaying it could apply the change twice.
	isRetriable := method == http.MethodGet || method == http.MethodHead || method == http.MethodOptions
	needsTokenRefresh := isRetriable && (resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden)
	if needsTokenRefresh {
		refreshCtx, cancelRefresh := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancelRefresh()
		accessToken, err = sessionProvider.RefreshSessionTokens(refreshCtx)
		if err != nil {
			return nil, api.StatusErrorf(http.StatusInternalServerError, "failed to refresh access token: %v", err)
		}

		err = resp.Body.Close()
		if err != nil {
			logger.Log.Warnw("Failed to close response body before retry", "error", err)
		}

		resp, err = sendInternalRequest(r, method, accessToken, userSecret, reqURL, rc, body)
		if err != nil {
			return nil, api.StatusErrorf(http.StatusInternalServerError, "failed to send request after token refresh: %v", err)
		}
	}

	return resp, nil
}

func sendInternalRequest(r *http.Request, method string, accessToken string, userSecret string, reqURL string, rc types.RouteConfig, body []byte) (*http.Response, error) {
	req, err := http.NewRequestWithContext(r.Context(), method, reqURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	if contentType := r.Header.Get("Content-Type"); contentType != "" {
		req.Header.Set("Content-Type", contentType)
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

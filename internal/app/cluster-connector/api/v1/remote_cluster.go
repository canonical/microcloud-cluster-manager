package v1

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/canonical/lxd/lxd/request"
	"github.com/canonical/lxd/lxd/response"
	"github.com/canonical/lxd/shared"
	"github.com/canonical/lxd/shared/api"
	"github.com/canonical/lxd/shared/version"
	"github.com/canonical/microcloud-cluster-manager/internal/app/cluster-connector/core/auth"
	"github.com/canonical/microcloud-cluster-manager/internal/app/cluster-connector/core/certificate"
	"github.com/canonical/microcloud-cluster-manager/internal/app/cluster-connector/core/cluster_link"
	"github.com/canonical/microcloud-cluster-manager/internal/app/cluster-connector/core/rate_limit"
	"github.com/canonical/microcloud-cluster-manager/internal/app/cluster-connector/core/tunnel"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/api/models/v1"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/config"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/database/store"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/logger"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/types"
	"github.com/golang/snappy"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/jmoiron/sqlx"
	"github.com/prometheus/common/expfmt"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/prompb"
)

// RemoteCluster is the API endpoint for managing remote clusters.
var RemoteCluster = types.RouteGroup{
	Prefix: "remote-cluster",
	Middlewares: []types.RouteMiddleware{
		rate_limit.RateLimitMiddleware,
	},
	Endpoints: []types.Endpoint{
		{
			Method:  http.MethodPost,
			Handler: remoteClustersPost,
		},
	},
}

// RemoteClusterProtected is the API endpoint for managing remote clusters with mtls authentication.
var RemoteClusterProtected = types.RouteGroup{
	Prefix: "remote-cluster",
	Middlewares: []types.RouteMiddleware{
		rate_limit.RateLimitMiddleware,
		auth.AuthMiddleware,
	},
	Endpoints: []types.Endpoint{
		{
			Path:    "status",
			Method:  http.MethodPost,
			Handler: remoteClusterStatusPost,
		},
		{
			Path:    "ws",
			Method:  http.MethodGet,
			Handler: remoteClusterWsGet,
		},
		{
			Method:  http.MethodDelete,
			Handler: remoteClusterDelete,
		},
	},
}

// RemoteClusterInternal are the API endpoints that cluster manager calls itself.
var RemoteClusterInternal = types.RouteGroup{
	Prefix: "remote-cluster",
	Middlewares: []types.RouteMiddleware{
		rate_limit.RateLimitMiddleware,
		auth.InternalAuthMiddleware,
	},
	Endpoints: []types.Endpoint{
		{
			Path:    "{remoteClusterName}/cluster-links",
			Method:  http.MethodGet,
			Handler: remoteClusterLinksGet,
		},
		{
			Path:    "{remoteClusterName}/cluster-links",
			Method:  http.MethodPost,
			Handler: remoteClusterLinksPost,
		},
		{
			Path:    "{remoteClusterName}/cluster-links/{clusterLinkName}",
			Method:  http.MethodDelete,
			Handler: remoteClusterLinkDelete,
		},
	},
}

func remoteClustersPost(rc types.RouteConfig) types.EndpointHandler {
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

		cert, err := certificate.ParseX509Certificate(payload.ClusterCertificate)
		if err != nil {
			logger.Log.Info("AUTHN invalid certificate on remote cluster post")
			return response.BadRequest(fmt.Errorf("invalid certificate: %v", err)).Render(w, r)
		}

		// get tokenFromDb secret for verification
		var tokenFromDb *store.RemoteClusterToken
		err = rc.DB.Transaction(r.Context(), func(ctx context.Context, tx *sqlx.Tx) error {
			var err error
			tokenFromDb, err = store.GetRemoteClusterToken(ctx, tx, payload.ClusterName)
			if err != nil {
				return err
			}

			return nil
		})

		if err != nil {
			logger.Log.Info("AUTHN could not find token in db on remote cluster post")
			return response.SmartError(err).Render(w, r)
		}

		// check if tokenFromDb has expired
		if time.Now().After(tokenFromDb.Expiry) {
			logger.Log.Info("AUTHN expired token used on remote cluster post")
			return response.Forbidden(fmt.Errorf("tokenFromDb has expired")).Render(w, r)
		}

		isTokenValid := strings.EqualFold(payload.Token, tokenFromDb.EncodedToken)
		if !isTokenValid {
			logger.Log.Info("AUTHN invalid token on remote cluster post")
			return response.Forbidden(err).Render(w, r)
		}

		// Create remote cluster entry and delete tokenFromDb in a single db transaction
		var remoteClusterID int
		err = rc.DB.Transaction(r.Context(), func(ctx context.Context, tx *sqlx.Tx) error {
			// create remote cluster entry
			newRemoteCluster, err := store.CreateRemoteCluster(ctx, tx, store.RemoteCluster{
				Name:               payload.ClusterName,
				Description:        tokenFromDb.Description,
				ClusterCertificate: payload.ClusterCertificate,
				JoinedAt:           time.Now(),
				Status:             string(models.ACTIVE),
			})

			if err != nil {
				return err
			}

			remoteClusterID = newRemoteCluster.ID

			// create relevant remote cluster details
			_, err = store.CreateRemoteClusterDetail(ctx, tx, store.RemoteClusterDetail{
				RemoteClusterID:   newRemoteCluster.ID,
				CephStatuses:      json.RawMessage("[]"),
				MemberStatuses:    json.RawMessage("[]"),
				InstanceStatuses:  json.RawMessage("[]"),
				StoragePoolUsages: json.RawMessage("[]"),
			})

			if err != nil {
				return err
			}

			_, err = store.CreateRemoteClusterConfig(ctx, tx, store.RemoteClusterConfig{
				RemoteClusterID: newRemoteCluster.ID,
				Key:             store.DiskThresholdKey,
				Value:           store.DiskThresholdDefault,
			})
			if err != nil {
				return err
			}

			_, err = store.CreateRemoteClusterConfig(ctx, tx, store.RemoteClusterConfig{
				RemoteClusterID: newRemoteCluster.ID,
				Key:             store.MemoryThresholdKey,
				Value:           store.MemoryThresholdDefault,
			})
			if err != nil {
				return err
			}

			// delete remote cluster tokenFromDb
			logger.Log.Info("AUTHN remote cluster join token consumed for new cluster")
			return store.DeleteRemoteClusterToken(ctx, tx, payload.ClusterName)
		})

		if err != nil {
			return response.SmartError(err).Render(w, r)
		}

		verifier, ok := rc.Auth.Authenticator.(*auth.MtlsAuthenticator)
		if ok {
			err = verifier.Cache().AddCertificate(cert.Certificate, remoteClusterID)
			if err != nil {
				return response.InternalError(err).Render(w, r)
			}
		}

		return response.EmptySyncResponse.Render(w, r)
	}
}

// mtls applied by the auth middleware for this endpoint.
func remoteClusterStatusPost(rc types.RouteConfig) types.EndpointHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		payload := models.RemoteClusterStatusPost{}
		err := json.NewDecoder(r.Body).Decode(&payload)
		if err != nil {
			logger.Log.Info("AUTHN invalid payload on remote cluster status post")
			return response.BadRequest(err).Render(w, r)
		}

		remoteClusterID, err := request.GetContextValue[int](r.Context(), auth.CtxRemoteClusterID)
		if err != nil {
			return response.SmartError(err).Render(w, r)
		}

		err = rc.DB.Transaction(r.Context(), func(ctx context.Context, tx *sqlx.Tx) error {
			dbRemoteCluster, err := store.GetRemoteClusterWithDetailByID(ctx, tx, remoteClusterID)
			if err != nil {
				return err
			}

			// Persist the cluster UUID reported in the heartbeat. Older MicroCloud
			// versions don't send it, so only update the stored value when a valid
			// UUID is provided and keep correlating via the certificate fingerprint.
			clusterUUID := dbRemoteCluster.ClusterUUID
			if payload.ClusterUUID != "" {
				_, err := uuid.Parse(payload.ClusterUUID)
				if err != nil {
					logger.Log.Warnw("Ignoring invalid cluster UUID in status update", "remote cluster", remoteClusterID, "cluster_uuid", payload.ClusterUUID)
				} else if dbRemoteCluster.ClusterUUID == "" {
					clusterUUID = payload.ClusterUUID
				} else if payload.ClusterUUID != dbRemoteCluster.ClusterUUID {
					logger.Log.Infow("Updating stored cluster UUID", "remote cluster", remoteClusterID, "old_cluster_uuid", dbRemoteCluster.ClusterUUID, "new_cluster_uuid", payload.ClusterUUID)
					clusterUUID = payload.ClusterUUID
				}
			}

			dbRemoteClusterDetail, err := store.GetRemoteClusterDetail(ctx, tx, remoteClusterID)
			if err != nil {
				return err
			}

			dbRemoteClusterDetail.Put(payload)
			err = store.UpdateRemoteClusterDetail(ctx, tx, dbRemoteCluster.ID, *dbRemoteClusterDetail)
			if err != nil {
				return err
			}

			newCluster := store.RemoteCluster{
				Name:               dbRemoteCluster.Name,
				Description:        dbRemoteCluster.Description,
				Status:             string(models.ACTIVE),
				JoinedAt:           time.Now(),
				ClusterCertificate: dbRemoteCluster.ClusterCertificate,
				ClusterUUID:        clusterUUID,
			}

			err = store.UpdateRemoteCluster(ctx, tx, dbRemoteCluster.Name, newCluster)
			if err != nil {
				return err
			}

			if len(payload.ServerMetrics) == 0 {
				return nil
			}

			if rc.Env.PrometheusBaseURL == "" || rc.Env.PrometheusBaseURL == "http://base" {
				logger.Log.Infow("Prometheus base URL is not configured, metrics received by cluster are discarded.", "remote cluster", remoteClusterID)
				return nil
			}

			for i := range payload.ServerMetrics {
				serverMetrics := payload.ServerMetrics[i]
				if serverMetrics.Service != "LXD" {
					logger.Log.Warnw("Unsupported service metrics received, skipping.", "service", serverMetrics.Service, "remote cluster", remoteClusterID)
					continue
				}

				timeSeries, err := parsePrometheusMetrics(serverMetrics.Metrics, dbRemoteCluster.Name)
				if err != nil {
					logger.Log.Warnw("Failed to parse metrics, skipping.", "remote cluster", remoteClusterID, "err", err)
					continue
				}

				err = forwardMetricsToPrometheus(timeSeries, rc)
				if err != nil {
					logger.Log.Warnw("Failed to forward metrics to Prometheus, skipping", "remote cluster", remoteClusterID, "err", err)
					continue
				}
			}

			return nil
		})

		if err != nil {
			logger.Log.Warnw("Failed to update remote cluster status", "remote cluster", remoteClusterID, "err", err)
			return response.SmartError(err).Render(w, r)
		}

		// TODO: determine next update time
		return response.SyncResponse(true, models.RemoteClusterStatusPostResponse{
			NextUpdateInSeconds:   time.Now().Local().String(),
			ClusterManagerAddress: rc.Env.ClusterConnectorDomain + ":" + rc.Env.ClusterConnectorPort,
		}).Render(w, r)
	}
}

// Parse the incoming Prometheus metrics (text format).
func parsePrometheusMetrics(metricsText string, jobName string) ([]prompb.TimeSeries, error) {
	parser := expfmt.NewTextParser(model.LegacyValidation)
	metricFamilies, err := parser.TextToMetricFamilies(bytes.NewReader([]byte(metricsText)))
	if err != nil {
		return nil, err
	}

	var timeSeries []prompb.TimeSeries
	for _, family := range metricFamilies {
		for _, sample := range family.GetMetric() {
			labels := make([]prompb.Label, 0, len(sample.GetLabel())+2)
			for _, label := range sample.GetLabel() {
				labels = append(labels, prompb.Label{
					Name:  label.GetName(),
					Value: label.GetValue(),
				})
			}

			// name of the metric e.g. lxd_cpu_seconds_total
			labels = append(labels, prompb.Label{Name: "__name__", Value: *family.Name})
			// allocate each metric to a job which maps to a remote cluster in this case
			labels = append(labels, prompb.Label{Name: "job", Value: strings.ReplaceAll(jobName, "-", "_")})

			// Get the metric value, lxd sends only counter and gauge metrics
			var metricValue float64
			if gauge := sample.GetGauge(); gauge != nil {
				metricValue = gauge.GetValue()
			} else if counter := sample.GetCounter(); counter != nil {
				metricValue = counter.GetValue()
			} else {
				return nil, fmt.Errorf("unsupported metric type for sample %v", sample)
			}

			timeSeries = append(timeSeries, prompb.TimeSeries{
				Labels: labels,
				Samples: []prompb.Sample{
					{
						Value:     metricValue,
						Timestamp: time.Now().UnixNano() / int64(time.Millisecond),
					},
				},
			})
		}
	}

	return timeSeries, nil
}

// Forward the metrics to Prometheus using remote-write.
func forwardMetricsToPrometheus(timeSeries []prompb.TimeSeries, rc types.RouteConfig) error {
	writeRequest := prompb.WriteRequest{
		Timeseries: timeSeries,
	}

	// Encode the WriteRequest as protobuf
	data, err := writeRequest.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal write request: %w", err)
	}

	// NOTE: prometheus requires the data to be compressed with snappy
	// ref: https://prometheus.io/docs/specs/remote_write_spec/
	compressedData := snappy.Encode(nil, data)

	remoteWriteURL := rc.Env.PrometheusBaseURL
	req, err := http.NewRequest("POST", remoteWriteURL, bytes.NewReader(compressedData))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("User-Agent", "microcloud-cluster-manager")
	req.Header.Set("X-Prometheus-Remote-Write-Version", "0.1.0")
	req.Header.Set("Content-Type", "application/x-protobuf")
	req.Header.Set("Content-Encoding", "snappy")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send metrics to Prometheus: %w", err)
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			logger.Log.Warnw("Failed to close Prometheus response body", "error", err)
		}
	}()

	if resp.StatusCode > 299 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read Prometheus response body: %w", err)
		}
		return fmt.Errorf("failed to send metrics to Prometheus, status: %s, response: %s", resp.Status, string(body))
	}

	return nil
}

// mtls applied by the auth middleware for this endpoint.
func remoteClusterDelete(rc types.RouteConfig) types.EndpointHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		remoteClusterID, err := request.GetContextValue[int](r.Context(), auth.CtxRemoteClusterID)
		if err != nil {
			return response.SmartError(err).Render(w, r)
		}

		err = rc.DB.Transaction(r.Context(), func(ctx context.Context, tx *sqlx.Tx) error {
			existing, err := store.GetRemoteClusterWithDetailByID(ctx, tx, remoteClusterID)
			if err != nil {
				return err
			}

			return store.DeleteRemoteCluster(ctx, tx, existing.Name)
		})

		if err != nil {
			return response.SmartError(err).Render(w, r)
		}

		return response.EmptySyncResponse.Render(w, r)
	}
}

// mtls applied by the auth middleware for this endpoint.
func remoteClusterWsGet(rc types.RouteConfig) types.EndpointHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		remoteClusterID, err := request.GetContextValue[int](r.Context(), auth.CtxRemoteClusterID)
		if err != nil {
			return response.SmartError(err).Render(w, r)
		}

		err = rc.DB.Transaction(r.Context(), func(ctx context.Context, tx *sqlx.Tx) error {
			_, err := store.GetRemoteClusterWithDetailByID(ctx, tx, remoteClusterID)
			if err != nil {
				return err
			}

			return nil
		})

		if err != nil {
			logger.Log.Warnw("Failed to verify remote cluster for websocket", "remote cluster", remoteClusterID, "err", err)
			return response.SmartError(err).Render(w, r)
		}

		var upgrader = websocket.Upgrader{}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			logger.Log.Warnw("Upgrade error", "error", err)
			return response.SmartError(err).Render(w, r)
		}

		tunnel := getClusterTunnel(rc, remoteClusterID)
		defer func() {
			logger.Log.Warnw("Client disconnected")
			tunnel.Mu.Lock()
			if tunnel.WsConn != conn {
				// someone else updated the tunnel already, do not close their connection
				tunnel.Mu.Unlock()
				return
			}

			tunnel.WsConn = nil
			tunnel.PendingResponses = make(map[string]chan types.ClusterManagerTunnelResponse)
			tunnel.Mu.Unlock()

			err := conn.Close()
			if err != nil {
				logger.Log.Warnw("Failed to close WebSocket connection", "error", err)
			}

			cleanupCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			err = rc.DB.Transaction(cleanupCtx, func(ctx context.Context, tx *sqlx.Tx) error {
				remoteCluster, err := store.GetRemoteClusterDetail(ctx, tx, remoteClusterID)
				if err != nil {
					return err
				}

				memberURL, err := getMemberURL(rc)
				if err != nil {
					return err
				}

				if remoteCluster.TunnelManagerMemberURL == memberURL {
					return nil // new tunnel was already established by another local member
				}

				// we are still registered as tunnel endpoint, clear it, because our tunnel was closed
				return store.UpdateRemoteClusterTunnel(ctx, tx, remoteClusterID, "")
			})
			if err != nil {
				logger.Log.Warnw("Failed to clear websocket member on disconnect", "remote cluster", remoteClusterID, "err", err)
			}
		}()

		logger.Log.Warnw("Client connected")

		// Close any previous connection
		ensureClosed(tunnel)

		tunnel.Mu.Lock()
		tunnel.WsConn = conn
		tunnel.Mu.Unlock()

		memberURL, err := getMemberURL(rc)
		if err != nil {
			return nil
		}

		err = rc.DB.Transaction(r.Context(), func(ctx context.Context, tx *sqlx.Tx) error {
			return store.UpdateRemoteClusterTunnel(ctx, tx, remoteClusterID, memberURL)
		})

		if err != nil {
			logger.Log.Warnw("Failed to store tunnel member url", "remote cluster", remoteClusterID, "err", err)
			return nil
		}

		for {
			var resp types.ClusterManagerTunnelResponse
			err := conn.ReadJSON(&resp)
			if err != nil {
				logger.Log.Errorf("WebSocket read error: %v", err)
				return nil
			}

			// Route the response to the waiting HTTP handler
			tunnel.Mu.RLock()
			ch, exists := tunnel.PendingResponses[resp.UUID]
			tunnel.Mu.RUnlock()

			if exists {
				select {
				case ch <- resp:
				default:
					logger.Log.Warnf("Dropping response for request ID %s: pending handler is no longer receiving", resp.UUID)
				}
			} else {
				logger.Log.Errorf("No pending request for response ID %s", resp.UUID)
			}
		}
	}
}

func getMemberURL(rc types.RouteConfig) (string, error) {
	memberIP, err := getMemberIP()
	if err != nil {
		return "", err
	}

	memberURL := "https://" + memberIP + ":" + rc.Env.ClusterConnectorPort
	return memberURL, nil
}

func getMemberIP() (string, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return "", err
	}

	ips, err := net.LookupIP(hostname)
	if err != nil {
		return "", err
	}

	for _, ip := range ips {
		if ip4 := ip.To4(); ip4 != nil {
			return ip4.String(), nil
		}
	}

	return "", fmt.Errorf("no IPv4 address found for hostname %q", hostname)
}

func ensureClosed(tunnel *types.Tunnel) {
	tunnel.Mu.Lock()
	conn := tunnel.WsConn
	if conn != nil {
		tunnel.WsConn = nil
		tunnel.PendingResponses = make(map[string]chan types.ClusterManagerTunnelResponse)
		tunnel.Mu.Unlock()

		err := conn.Close()
		if err != nil {
			logger.Log.Warnw("Failed to close existing WebSocket connection", "error", err)
		}
	} else {
		tunnel.Mu.Unlock()
	}
}

func getClusterTunnel(rc types.RouteConfig, remoteClusterID int) *types.Tunnel {
	rc.TunnelStore.Mu.Lock()
	defer rc.TunnelStore.Mu.Unlock()

	tunnel := rc.TunnelStore.TunnelByCluster[remoteClusterID]
	if tunnel == nil {
		// Initialize a new tunnel for this cluster
		tunnel = &types.Tunnel{
			WsConn:           nil, // will be set.
			PendingResponses: make(map[string]chan types.ClusterManagerTunnelResponse),
			UserSessions:     make(map[string]string),
		}

		rc.TunnelStore.TunnelByCluster[remoteClusterID] = tunnel
	}

	return tunnel
}

func remoteClusterLinksPost(rc types.RouteConfig) types.EndpointHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		remoteClusterName, err := url.PathUnescape(mux.Vars(r)["remoteClusterName"])
		if err != nil {
			return response.SmartError(err).Render(w, r)
		}

		body, err := readBody(r)
		if err != nil {
			logger.Log.Errorw("Failed to read request body", "error", err)
			return response.SmartError(errors.New("Failed to read request body")).Render(w, r)
		}

		var linkRequest models.RemoteClusterLinkPost
		err = json.Unmarshal(body, &linkRequest)
		if err != nil {
			logger.Log.Errorw("Failed to parse cluster link request body", "error", err, "body_length", len(body))
			return response.BadRequest(fmt.Errorf("Failed to parse request body: %w", err)).Render(w, r)
		}

		if linkRequest.TargetCluster == "" {
			return response.BadRequest(errors.New("Missing target cluster")).Render(w, r)
		}

		if linkRequest.TargetCluster == remoteClusterName {
			return response.BadRequest(errors.New("Source cluster and target cluster must differ")).Render(w, r)
		}

		if linkRequest.Type == cluster_link.TypePublic && len(linkRequest.AuthGroups) > 0 {
			return response.BadRequest(errors.New("Auth groups cannot be set for public cluster links")).Render(w, r)
		}

		if linkRequest.Type != cluster_link.TypePublic && linkRequest.TargetAddress != "" {
			return response.BadRequest(errors.New("Target address can only be set for public cluster links")).Render(w, r)
		}

		err = checkClusterLinkPreflight(rc, r, linkRequest.Type, remoteClusterName, linkRequest.TargetCluster)
		if err != nil {
			return response.SmartError(err).Render(w, r)
		}

		// The link created on the source cluster is named after the target cluster unless a name is given.
		if linkRequest.Name == "" {
			linkRequest.Name = linkRequest.TargetCluster
		}

		switch linkRequest.Type {
		case cluster_link.TypeUnidirectional:
			return createUnidirectionalClusterLink(rc, w, r, remoteClusterName, linkRequest)
		case cluster_link.TypePublic:
			return createPublicClusterLink(rc, w, r, remoteClusterName, linkRequest)
		default:
			return createBidirectionalClusterLink(rc, w, r, remoteClusterName, linkRequest)
		}
	}
}

const clusterLinksPath = "/1.0/cluster/links"

// isSuccessStatus reports whether the status is a 2xx status code.
func isSuccessStatus(status int) bool {
	return status >= http.StatusOK && status < http.StatusMultipleChoices
}

func sendClusterLinkRequest(rc types.RouteConfig, r *http.Request, clusterName string, requestBody any) (*types.ClusterManagerTunnelResponse, error) {
	const operation = "create cluster link"
	body, err := json.Marshal(requestBody)
	if err != nil {
		logger.Log.Errorw("Failed to encode cluster link request", "cluster", clusterName, "error", err)
		return nil, api.StatusErrorf(http.StatusInternalServerError, "Failed to encode cluster link request for cluster %q", clusterName)
	}

	resp, err := sendTunnelRequest(rc, r, http.MethodPost, clusterLinksPath, clusterName, body)
	if err != nil {
		logger.Log.Errorw("Cluster link request failed", "cluster", clusterName, "operation", operation, "error", err)
		if api.StatusErrorCheck(err, http.StatusUnauthorized, http.StatusForbidden) {
			return nil, err
		}
		if api.StatusErrorCheck(err, http.StatusNotFound) {
			return nil, fmt.Errorf("Failed to %s on cluster %q: %w", operation, clusterName, err)
		}
		return nil, api.StatusErrorf(http.StatusBadGateway, "Failed to %s on cluster %q: %v", operation, clusterName, err)
	}
	if !isSuccessStatus(resp.Status) {
		err = tunnel.ResponseError(*resp, clusterName, operation)
		logger.Log.Errorw("Cluster link request failed", "cluster", clusterName, "operation", operation, "status", resp.Status, "error", err)
		return nil, err
	}
	return resp, nil
}

func extractCertificateFingerprint(body []byte) (string, error) {
	var certificateResponse struct {
		Metadata api.Certificate `json:"metadata"`
	}
	if err := json.Unmarshal(body, &certificateResponse); err != nil {
		return "", fmt.Errorf("failed to parse certificate response: %w", err)
	}
	if certificateResponse.Metadata.Fingerprint == "" {
		return "", errors.New("certificate response did not contain a fingerprint")
	}
	return certificateResponse.Metadata.Fingerprint, nil
}

// createBidirectionalClusterLink creates a pending link on the source cluster and activates it on the target cluster
// using the trust token returned by the source cluster.
func createBidirectionalClusterLink(rc types.RouteConfig, w http.ResponseWriter, r *http.Request, remoteClusterName string, linkRequest models.RemoteClusterLinkPost) error {
	sourceLinkName := linkRequest.Name

	// Create the pending cluster link on the source cluster, which returns a trust token.
	pendingLinkResponse, err := sendClusterLinkRequest(rc, r, remoteClusterName, linkRequest)
	if err != nil {
		return response.SmartError(err).Render(w, r)
	}

	activated := false
	defer func() {
		if !activated {
			deletePendingClusterLink(rc, r, remoteClusterName, sourceLinkName)
		}
	}()

	trustToken, err := extractTrustToken(pendingLinkResponse.Body)
	if err != nil {
		logger.Log.Errorw("Failed to extract trust token from cluster link response", "error", err)
		return response.SmartError(api.StatusErrorf(http.StatusBadGateway, "Failed to extract trust token from cluster link response")).Render(w, r)
	}

	// Activate the cluster link on the target cluster using the trust token.
	targetRequest := api.ClusterLinksPost{
		ClusterLinkPut: api.ClusterLinkPut{
			Description: linkRequest.Description,
		},
		Name:       remoteClusterName,
		Type:       linkRequest.Type,
		TrustToken: trustToken,
		AuthGroups: linkRequest.AuthGroups,
	}

	targetLinkResponse, err := sendClusterLinkRequest(rc, r, linkRequest.TargetCluster, targetRequest)
	if err != nil {
		return response.SmartError(err).Render(w, r)
	}

	activated = true
	writeResponse(w, *targetLinkResponse)

	return nil
}

// createUnidirectionalClusterLink creates a pending identity on the target cluster and uses the returned trust token to create a link on the source cluster.
// The target cluster does not get a link to the source.
func createUnidirectionalClusterLink(rc types.RouteConfig, w http.ResponseWriter, r *http.Request, remoteClusterName string, linkRequest models.RemoteClusterLinkPost) error {
	// Issue the pending identity on the target cluster. It is named after the source cluster.
	identityRequest := api.ClusterLinksPost{
		Name:       remoteClusterName,
		Type:       cluster_link.TypeUnidirectional,
		AuthGroups: linkRequest.AuthGroups,
	}

	identityResponse, err := sendClusterLinkRequest(rc, r, linkRequest.TargetCluster, identityRequest)
	if err != nil {
		return response.SmartError(err).Render(w, r)
	}

	activated := false
	defer func() {
		if !activated {
			deletePendingIdentity(rc, r, linkRequest.TargetCluster, remoteClusterName)
		}
	}()

	trustToken, err := extractTrustToken(identityResponse.Body)
	if err != nil {
		logger.Log.Errorw("Failed to extract trust token from pending identity response", "error", err)
		return response.SmartError(api.StatusErrorf(http.StatusBadGateway, "Failed to extract trust token from pending identity response")).Render(w, r)
	}

	// Create the link on the source cluster, which also activates the pending identity on the target cluster.
	sourceRequest := api.ClusterLinksPost{
		ClusterLinkPut: api.ClusterLinkPut{
			Description: linkRequest.Description,
		},
		Name:       linkRequest.Name,
		Type:       cluster_link.TypeUnidirectional,
		TrustToken: trustToken,
	}

	sourceResponse, err := sendClusterLinkRequest(rc, r, remoteClusterName, sourceRequest)
	if err != nil {
		return response.SmartError(err).Render(w, r)
	}

	activated = true
	writeResponse(w, *sourceResponse)

	return nil
}

// publicClusterLinkPost extends the LXD cluster link request with the fields used by the two phases of public
// cluster link creation.
type publicClusterLinkPost struct {
	api.ClusterLinksPost `json:",inline"`

	Fingerprint   string `json:"fingerprint,omitempty"`
	RemoteAddress string `json:"remote_address,omitempty"`
}

// createPublicClusterLink creates a link on the source cluster that can reach the target cluster without any
// authentication. The source cluster and cluster manager independently fetch the target certificate; the UI lets
// the user verify its fingerprint before the link is activated.
func createPublicClusterLink(rc types.RouteConfig, w http.ResponseWriter, r *http.Request, remoteClusterName string, linkRequest models.RemoteClusterLinkPost) error {
	publicRequest := publicClusterLinkPost{
		ClusterLinksPost: api.ClusterLinksPost{
			ClusterLinkPut: api.ClusterLinkPut{
				Description: linkRequest.Description,
			},
			Name: linkRequest.Name,
			Type: cluster_link.TypePublic,
		},
	}

	// If the fingerprint is provided, we are in the confirmation phase and can directly send the request.
	if linkRequest.Fingerprint != "" {
		publicRequest.Fingerprint = linkRequest.Fingerprint
		confirmResponse, err := sendClusterLinkRequest(rc, r, remoteClusterName, publicRequest)
		if err != nil {
			return response.SmartError(err).Render(w, r)
		}

		writeResponse(w, *confirmResponse)
		return nil
	}

	targetAddressForSourceCluster, registeredTargetURLForClusterManager, err := getPublicClusterLinkAddresses(rc, r, linkRequest.TargetCluster, linkRequest.TargetAddress)
	if err != nil {
		return response.SmartError(err).Render(w, r)
	}
	parsedRegisteredTargetURL, err := url.Parse(registeredTargetURLForClusterManager)
	if err != nil || parsedRegisteredTargetURL.Scheme != "https" || parsedRegisteredTargetURL.Host == "" || parsedRegisteredTargetURL.User != nil {
		return response.SmartError(api.StatusErrorf(http.StatusBadRequest, "Cluster %q has an invalid registered HTTPS LXD address", linkRequest.TargetCluster)).Render(w, r)
	}

	certificateCtx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	targetCert, err := shared.GetRemoteCertificateNoProxy(certificateCtx, registeredTargetURLForClusterManager, version.UserAgent)
	if err != nil {
		logger.Log.Errorw("Failed to fetch target LXD certificate", "target_cluster", linkRequest.TargetCluster, "error", err)
		return response.SmartError(api.StatusErrorf(http.StatusBadGateway, "Failed to fetch the LXD certificate of cluster %q", linkRequest.TargetCluster)).Render(w, r)
	}
	targetFingerprint := shared.CertFingerprint(targetCert)

	publicRequest.RemoteAddress = targetAddressForSourceCluster

	// Phase one: create the pending link, the source cluster fetches the target certificate.
	pendingResponse, err := sendClusterLinkRequest(rc, r, remoteClusterName, publicRequest)
	if err != nil {
		return response.SmartError(err).Render(w, r)
	}

	keepPending := false
	defer func() {
		if !keepPending {
			deletePendingClusterLink(rc, r, remoteClusterName, linkRequest.Name)
		}
	}()

	fingerprint, err := extractCertificateFingerprint(pendingResponse.Body)
	if err != nil {
		logger.Log.Errorw("Failed to read the certificate fingerprint of the pending public cluster link", "error", err)
		return response.SmartError(api.StatusErrorf(http.StatusBadGateway, "Failed to read the certificate fingerprint of the pending public cluster link")).Render(w, r)
	}

	if fingerprint != targetFingerprint {
		logger.Log.Errorw("Source cluster and cluster manager observed different certificates for public cluster link", "target_cluster", linkRequest.TargetCluster, "address", targetAddressForSourceCluster)
		return response.SmartError(api.StatusErrorf(http.StatusBadGateway, "Source cluster and cluster manager observed different certificates from cluster %q", linkRequest.TargetCluster)).Render(w, r)
	}

	keepPending = true
	return response.SyncResponse(true, models.RemoteClusterLinkCertificate{Fingerprint: fingerprint}).Render(w, r)
}

// getPublicClusterLinkAddresses returns the target address for the source cluster and registered target URL for the cluster manager.
func getPublicClusterLinkAddresses(rc types.RouteConfig, r *http.Request, clusterName string, targetAddressOverrideForSourceCluster string) (targetAddressForSourceCluster string, registeredTargetURLForClusterManager string, err error) {
	var details *store.RemoteClusterWithDetail
	err = rc.DB.Transaction(r.Context(), func(ctx context.Context, tx *sqlx.Tx) error {
		details, err = store.GetRemoteClusterWithDetailByName(ctx, tx, clusterName)
		return err
	})
	if err != nil {
		return "", "", err
	}
	if details == nil || strings.TrimSpace(details.LXDURL) == "" {
		return "", "", api.StatusErrorf(http.StatusBadRequest, "Cluster %q has no registered LXD address for certificate verification", clusterName)
	}

	registeredTargetURLForClusterManager = strings.TrimSpace(details.LXDURL)
	targetURLForSourceCluster := registeredTargetURLForClusterManager
	if strings.TrimSpace(targetAddressOverrideForSourceCluster) != "" {
		targetURLForSourceCluster = strings.TrimSpace(targetAddressOverrideForSourceCluster)
	}

	parsedTargetURLForSourceCluster, err := url.Parse(targetURLForSourceCluster)
	if err != nil || parsedTargetURLForSourceCluster.Scheme != "https" || parsedTargetURLForSourceCluster.Host == "" || parsedTargetURLForSourceCluster.User != nil {
		return "", "", api.StatusErrorf(http.StatusBadRequest, "Target address for cluster %q must be a valid HTTPS URL", clusterName)
	}

	return parsedTargetURLForSourceCluster.Host, registeredTargetURLForClusterManager, nil
}

// deletePendingIdentity removes a pending identity issued on the target cluster of a unidirectional link.
// Failures are logged only, as the caller is already handling an error.
func deletePendingIdentity(rc types.RouteConfig, r *http.Request, targetClusterName string, identity string) {
	ctx, cancel := context.WithTimeout(context.WithoutCancel(r.Context()), 5*time.Second)
	defer cancel()
	r = r.WithContext(ctx)

	path := "/1.0/auth/identities/tls/" + url.PathEscape(identity)
	resp, err := sendTunnelRequest(rc, r, http.MethodDelete, path, targetClusterName, nil)
	if err != nil {
		logger.Log.Errorw("Failed to remove pending identity", "target_cluster", targetClusterName, "identity", identity, "error", err)
		return
	}

	if !isSuccessStatus(resp.Status) {
		err = tunnel.ResponseError(*resp, targetClusterName, "remove pending identity")
		logger.Log.Errorw("Failed to remove pending identity", "target_cluster", targetClusterName, "identity", identity, "status", resp.Status, "error", err)
	}
}

// deletePendingClusterLink removes a cluster link that was created on the source cluster but could not be
// activated on the target cluster. Failures are logged only, as the caller is already handling an error.
func deletePendingClusterLink(rc types.RouteConfig, r *http.Request, sourceClusterName string, clusterLinkName string) {
	if clusterLinkName == "" {
		return
	}

	ctx, cancel := context.WithTimeout(context.WithoutCancel(r.Context()), 5*time.Second)
	defer cancel()
	r = r.WithContext(ctx)

	path := "/1.0/cluster/links/" + url.PathEscape(clusterLinkName)
	resp, err := sendTunnelRequest(rc, r, http.MethodDelete, path, sourceClusterName, nil)
	if err != nil {
		logger.Log.Errorw("Failed to remove pending cluster link", "source_cluster", sourceClusterName, "cluster_link", clusterLinkName, "error", err)
		return
	}

	if !isSuccessStatus(resp.Status) {
		err = tunnel.ResponseError(*resp, sourceClusterName, "remove pending cluster link")
		logger.Log.Errorw("Failed to remove pending cluster link", "source_cluster", sourceClusterName, "cluster_link", clusterLinkName, "status", resp.Status, "error", err)
	}
}

// extractTrustToken reads the trust token returned by LXD when creating a pending cluster link
// and returns it as a base64 encoded string, ready to be used as the trust_token of a cluster link request.
func extractTrustToken(body []byte) (string, error) {
	var lxdResponse struct {
		Metadata api.CertificateAddToken `json:"metadata"`
	}

	err := json.Unmarshal(body, &lxdResponse)
	if err != nil {
		return "", fmt.Errorf("failed to parse cluster link response: %w", err)
	}

	if lxdResponse.Metadata.Secret == "" || lxdResponse.Metadata.Fingerprint == "" {
		return "", errors.New("cluster link response did not contain a trust token")
	}

	token := lxdResponse.Metadata.String()
	if token == "" {
		return "", errors.New("failed to encode trust token")
	}

	return token, nil
}

// checkClusterLinkPreflight validates the link type and verifies that both clusters meet its
// API-extension and permission requirements before a link is created.
func checkClusterLinkPreflight(rc types.RouteConfig, r *http.Request, linkType string, sourceCluster string, targetCluster string) error {
	requirements, ok := cluster_link.RequirementsFor(linkType)
	if !ok {
		return api.StatusErrorf(http.StatusBadRequest, "Unsupported cluster link type %q", linkType)
	}

	var sourceErr, targetErr error
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		sourceErr = checkTunnelRequirements(rc, r, sourceCluster, requirements.Source)
	}()
	go func() {
		defer wg.Done()
		targetErr = checkTunnelRequirements(rc, r, targetCluster, requirements.Target)
	}()
	wg.Wait()

	if api.StatusErrorCheck(sourceErr, http.StatusUnauthorized, http.StatusForbidden) {
		return sourceErr
	}
	if api.StatusErrorCheck(targetErr, http.StatusUnauthorized, http.StatusForbidden) {
		return targetErr
	}

	if sourceErr != nil {
		if targetErr != nil {
			return api.StatusErrorf(http.StatusBadGateway, "Source cluster error: %v; Target cluster error: %v", sourceErr, targetErr)
		}

		return sourceErr
	}

	return targetErr
}

// checkTunnelRequirements verifies that a cluster is reachable over its tunnel, supports the
// requested API extensions, and grants the user the given entity-qualified permissions. An empty
// permissions list means only reachability and API-extension support are checked.
func checkTunnelRequirements(rc types.RouteConfig, r *http.Request, clusterName string, requirements tunnel.ClusterRequirements) error {
	// Probe reachability and required API-extension support.
	resp, err := sendTunnelRequest(rc, r, http.MethodGet, "/1.0", clusterName, nil)
	if err != nil {
		// Do not mask a 404 Not Found error as a bad gateway.
		if api.StatusErrorCheck(err, http.StatusNotFound, http.StatusUnauthorized, http.StatusForbidden) {
			return err
		}
		return api.StatusErrorf(http.StatusBadGateway, "Cluster %q is unreachable: %s", clusterName, err)
	}

	if resp.Status < http.StatusOK || resp.Status >= http.StatusMultipleChoices {
		err = tunnel.ResponseError(*resp, clusterName, "check cluster capabilities")
		logger.Log.Errorw("Cluster capability check failed", "cluster", clusterName, "error", err)
		return err
	}

	var serverResponse struct {
		Metadata api.Server `json:"metadata"`
	}

	err = json.Unmarshal(resp.Body, &serverResponse)
	if err != nil {
		logger.Log.Errorw("Failed to parse cluster server response", "cluster", clusterName, "error", err)
		return api.StatusErrorf(http.StatusBadGateway, "Invalid server response from cluster %q", clusterName)
	}

	for _, apiExtension := range requirements.APIExtensions {
		if !slices.Contains(serverResponse.Metadata.APIExtensions, apiExtension) {
			return api.StatusErrorf(http.StatusBadRequest, "Cluster %q does not support required API extension %q", clusterName, apiExtension)
		}
	}

	if len(requirements.Permissions) == 0 || !slices.Contains(serverResponse.Metadata.APIExtensions, tunnel.AccessManagementAPIExtension) {
		// Nothing more to check, or the cluster cannot report effective permissions (older LXD);
		// any missing permissions will surface as creation-time errors.
		return nil
	}

	// Let LXD evaluate the user's effective permissions.
	resp, err = sendTunnelRequest(rc, r, http.MethodGet, "/1.0/auth/identities/current", clusterName, nil)
	if err != nil {
		if api.StatusErrorCheck(err, http.StatusUnauthorized, http.StatusForbidden) {
			return err
		}
		return api.StatusErrorf(http.StatusBadGateway, "Cluster %q is unreachable: %s", clusterName, err)
	}

	if resp.Status < http.StatusOK || resp.Status >= http.StatusMultipleChoices {
		err = tunnel.ResponseError(*resp, clusterName, "check cluster permissions")
		logger.Log.Errorw("Cluster permission check failed", "cluster", clusterName, "error", err)
		return err
	}

	var identityResponse struct {
		Metadata api.IdentityInfo `json:"metadata"`
	}

	err = json.Unmarshal(resp.Body, &identityResponse)
	if err != nil {
		logger.Log.Errorw("Failed to parse cluster identity response", "cluster", clusterName, "error", err)
		return api.StatusErrorf(http.StatusBadGateway, "Invalid identity response from cluster %q", clusterName)
	}

	if !identityResponse.Metadata.FineGrained {
		// Unrestricted identities (e.g. trusted TLS clients) hold every permission.
		return nil
	}

	err = tunnel.CheckPermissions(identityResponse.Metadata.EffectivePermissions, requirements.Permissions)
	if err != nil {
		return api.StatusErrorf(http.StatusForbidden, "Insufficient permissions on cluster %q: %v", clusterName, err)
	}

	return nil
}

func remoteClusterLinksGet(rc types.RouteConfig) types.EndpointHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		remoteClusterName, err := url.PathUnescape(mux.Vars(r)["remoteClusterName"])
		if err != nil {
			return response.SmartError(err).Render(w, r)
		}

		body, err := readBody(r)
		if err != nil {
			logger.Log.Errorw("Failed to read request body", "error", err)
			return response.SmartError(errors.New("Failed to read request body")).Render(w, r)
		}

		path := "/1.0/cluster/links?recursion=2"
		resp, err := sendTunnelRequest(rc, r, http.MethodGet, path, remoteClusterName, body)
		if err != nil {
			return err
		}
		if !isSuccessStatus(resp.Status) {
			err = tunnel.ResponseError(*resp, remoteClusterName, "list cluster links")
			logger.Log.Errorw("Failed to list cluster links", "cluster", remoteClusterName, "status", resp.Status, "error", err)
			return response.SmartError(err).Render(w, r)
		}
		writeResponse(w, *resp)
		return nil
	}
}

// remoteClusterLinkDelete removes a single edge of a cluster link, meaning the link is only removed on the
// remote cluster given in the path. The linked cluster keeps its own cluster link entry.
func remoteClusterLinkDelete(rc types.RouteConfig) types.EndpointHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		remoteClusterName, err := url.PathUnescape(mux.Vars(r)["remoteClusterName"])
		if err != nil {
			return response.SmartError(err).Render(w, r)
		}

		clusterLinkName, err := url.PathUnescape(mux.Vars(r)["clusterLinkName"])
		if err != nil {
			return response.SmartError(err).Render(w, r)
		}

		if clusterLinkName == "" {
			return response.BadRequest(errors.New("Missing cluster link name")).Render(w, r)
		}

		path := "/1.0/cluster/links/" + url.PathEscape(clusterLinkName)
		resp, err := sendTunnelRequest(rc, r, http.MethodDelete, path, remoteClusterName, nil)
		if err != nil {
			return response.SmartError(err).Render(w, r)
		}

		if !isSuccessStatus(resp.Status) {
			err = tunnel.ResponseError(*resp, remoteClusterName, "delete cluster link")
			logger.Log.Errorw("Failed to delete cluster link", "cluster", remoteClusterName, "cluster_link", clusterLinkName, "status", resp.Status, "error", err)
			return response.SmartError(err).Render(w, r)
		}

		writeResponse(w, *resp)
		return nil
	}
}

func sendTunnelRequest(rc types.RouteConfig, r *http.Request, method string, path string, remoteClusterName string, body []byte) (*types.ClusterManagerTunnelResponse, error) {
	authorizationHeader := r.Header.Get("Authorization")
	if authorizationHeader == "" {
		return nil, errors.New("Authorization header is missing")
	}

	userSecret := r.Header.Get("X-User-Secret")
	if userSecret == "" {
		return nil, errors.New("X-User-Secret header is missing")
	}

	var remoteClusterID int
	var err error
	err = rc.DB.Transaction(r.Context(), func(ctx context.Context, tx *sqlx.Tx) error {
		remoteClusterID, err = store.GetRemoteClusterID(ctx, tx, remoteClusterName)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	id, err := uuid.NewV7()
	if err != nil {
		logger.Log.Errorw("Failed to generate request ID", "error", err)
		return nil, errors.New("Failed to generate request ID")
	}

	rc.TunnelStore.Mu.Lock()
	tunnel := rc.TunnelStore.TunnelByCluster[remoteClusterID]
	rc.TunnelStore.Mu.Unlock()

	if tunnel == nil {
		return nil, errors.New("Tunnel not found")
	}

	headers := http.Header{}
	headers.Set("Authorization", authorizationHeader)
	session, err := getUserSession(tunnel, authorizationHeader, userSecret)
	if err != nil {
		return nil, errors.New("Error loading LXD session value")
	}

	if session != "" {
		headers.Set("Cookie", "session="+session)
	}

	req := types.ClusterManagerTunnelRequest{
		UUID:    id.String(),
		Method:  method,
		Path:    path,
		Headers: headers,
		Body:    body,
	}

	// register for response
	ch := make(chan types.ClusterManagerTunnelResponse, 1)
	tunnel.Mu.Lock()
	tunnel.PendingResponses[id.String()] = ch
	tunnel.Mu.Unlock()
	defer func() {
		tunnel.Mu.Lock()
		delete(tunnel.PendingResponses, id.String())
		tunnel.Mu.Unlock()
	}()

	// send request
	tunnel.Mu.Lock()
	wsConn := tunnel.WsConn
	if wsConn == nil {
		tunnel.Mu.Unlock()
		return nil, errors.New("Tunnel not connected")
	}
	err = wsConn.WriteJSON(req)
	tunnel.Mu.Unlock()

	if err != nil {
		logger.Log.Errorw("Failed to send request over WebSocket", "error", err)
		ensureClosed(tunnel)
		return nil, errors.New("Failed to send request")
	}

	// Wait for response
	select {
	case resp := <-ch:
		setUserSession(resp, tunnel, authorizationHeader, userSecret)
		return &resp, nil
	case <-time.After(15 * time.Second):
		return nil, errors.New("Timeout")
	case <-r.Context().Done():
		return nil, errors.New("Client disconnected")
	}
}

func readBody(r *http.Request) ([]byte, error) {
	defer func() {
		err := r.Body.Close()
		if err != nil {
			logger.Log.Warnw("Failed to close request body", "error", err)
		}
	}()
	return io.ReadAll(r.Body)
}

func writeResponse(w http.ResponseWriter, resp types.ClusterManagerTunnelResponse) {
	for k, values := range resp.Headers {
		w.Header()[k] = append([]string(nil), values...)
	}

	w.WriteHeader(resp.Status)
	_, err := w.Write(resp.Body)
	if err != nil {
		logger.Log.Errorf("error writing response body: %v", err)
	}
}

func getUserSession(tunnel *types.Tunnel, authorizationToken string, userSecret string) (string, error) {
	tunnel.Mu.Lock()
	encryptedSession, hasSession := tunnel.UserSessions[authorizationToken]
	tunnel.Mu.Unlock()
	if !hasSession {
		return "", nil
	}

	session, err := config.DecryptUserValue(encryptedSession, userSecret)
	if err != nil {
		return "", err
	}

	return session, nil
}

func setUserSession(resp types.ClusterManagerTunnelResponse, tunnel *types.Tunnel, authorizationToken string, userSecret string) {
	var sessionID string
	for _, cookie := range resp.Cookies {
		if cookie.Name == "session" {
			sessionID = cookie.Value
		}
	}

	if sessionID == "" {
		logger.Log.Warnf("Failed to read session value from tunnel response, cookie not found")
		return
	}

	encryptedSessionID, err := config.EncryptUserValue(sessionID, userSecret)
	if err != nil {
		logger.Log.Errorf("Failed to encrypt session value: %v", err)
		return
	}

	tunnel.Mu.Lock()
	tunnel.UserSessions[authorizationToken] = encryptedSessionID
	tunnel.Mu.Unlock()
}

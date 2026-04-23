package v1

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/canonical/lxd/lxd/request"
	"github.com/canonical/lxd/lxd/response"
	"github.com/canonical/microcloud-cluster-manager/internal/app/cluster-connector/core/auth"
	"github.com/canonical/microcloud-cluster-manager/internal/app/cluster-connector/core/certificate"
	"github.com/canonical/microcloud-cluster-manager/internal/app/cluster-connector/core/rate_limit"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/api/models/v1"
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

// apply mtls for this endpoint.
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

			if dbRemoteCluster == nil {
				return fmt.Errorf("remote cluster not found")
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

// apply mtls for this endpoint.
func remoteClusterWsGet(rc types.RouteConfig) types.EndpointHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		remoteClusterID, err := request.GetContextValue[int](r.Context(), auth.CtxRemoteClusterID)
		if err != nil {
			return response.SmartError(err).Render(w, r)
		}

		err = rc.DB.Transaction(r.Context(), func(ctx context.Context, tx *sqlx.Tx) error {
			dbRemoteCluster, err := store.GetRemoteClusterWithDetailByID(ctx, tx, remoteClusterID)
			if err != nil {
				return err
			}

			if dbRemoteCluster == nil {
				return fmt.Errorf("remote cluster not found")
			}

			// todo: use dynamic url of the current host
			memberUrl := "https://localhost:9000"
			return storeTunnelMemberUrl(ctx, tx, remoteClusterID, memberUrl)
		})

		if err != nil {
			logger.Log.Warnw("Failed to verify remote cluster for websocket", "remote cluster", remoteClusterID, "err", err)
			return response.SmartError(err).Render(w, r)
		}

		var upgrader = websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // allow all origins
			},
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			logger.Log.Warnw("Upgrade error:", err)
			return response.SmartError(err).Render(w, r)
		}

		tunnel := getClusterTunnel(rc, remoteClusterID)
		defer func() {
			logger.Log.Warnw("Client disconnected")
			ensureClosed(tunnel)
			err = rc.DB.Transaction(context.Background(), func(ctx context.Context, tx *sqlx.Tx) error {
				return storeTunnelMemberUrl(ctx, tx, remoteClusterID, "")
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

		for {
			var resp types.ClusterManagerTunnelResponse
			err := conn.ReadJSON(&resp)
			if err != nil {
				logger.Log.Error("WebSocket read error: %v", err)
				return nil
			}

			// Route the response to the waiting HTTP handler
			tunnel.Mu.RLock()
			ch, exists := tunnel.PendingCalls[resp.ID]
			tunnel.Mu.RUnlock()

			if exists {
				ch <- resp
			} else {
				logger.Log.Error("No pending request for response ID %s", resp.ID)
			}
		}
	}
}

func ensureClosed(tunnel *types.Tunnel) {
	conn := tunnel.WsConn
	if conn != nil {
		tunnel.Mu.Lock()
		tunnel.WsConn = nil
		tunnel.Mu.Unlock()

		err := conn.Close()
		if err != nil {
			logger.Log.Warnw("Failed to close existing WebSocket connection", "error", err)
		}
	}
}

func getClusterTunnel(rc types.RouteConfig, remoteClusterID int) *types.Tunnel {
	rc.TunnelStore.Mu.Lock()
	defer rc.TunnelStore.Mu.Unlock()

	tunnel := rc.TunnelStore.TunnelByCluster[remoteClusterID]
	if tunnel == nil {
		// Initialize a new tunnel for this cluster
		tunnel = &types.Tunnel{
			Mu:           sync.RWMutex{},
			WsConn:       nil, // will be set.
			PendingCalls: make(map[string]chan types.ClusterManagerTunnelResponse),
		}

		rc.TunnelStore.TunnelByCluster[remoteClusterID] = tunnel
	}

	return tunnel
}

func storeTunnelMemberUrl(ctx context.Context, tx *sqlx.Tx, remoteClusterID int, memberUrl string) error {
	dbRemoteClusterDetail, err := store.GetRemoteClusterDetail(ctx, tx, remoteClusterID)
	if err != nil {
		return err
	}

	dbRemoteClusterDetail.TunnelMemberURL = memberUrl
	err = store.UpdateRemoteClusterDetail(ctx, tx, remoteClusterID, *dbRemoteClusterDetail)
	if err != nil {
		return err
	}
	return nil
}

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

func remoteClusterTunnel(rc types.RouteConfig) types.EndpointHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		remoteClusterName, err := url.PathUnescape(mux.Vars(r)["remoteClusterName"])
		if err != nil {
			return response.SmartError(err).Render(w, r)
		}

		var remoteClusterID int
		err = rc.DB.Transaction(r.Context(), func(ctx context.Context, tx *sqlx.Tx) error {
			remoteCluster, err := store.GetRemoteCluster(ctx, tx, remoteClusterName)
			if err != nil {
				return err
			}

			remoteClusterID = remoteCluster.ID

			return nil
		})
		if err != nil {
			return response.SmartError(err).Render(w, r)
		}

		id := uuid.New().String()
		body, err := readBody(r)
		if err != nil {
			logger.Log.Errorw("Failed to read request body", "error", err)
			http.Error(w, "Failed to read request body", http.StatusInternalServerError)
			return nil
		}

		logger.Log.Debug("remote cluster tunnel request", "id", id, "method", r.Method, "path", r.URL.Path)

		rc.TunnelStore.Mu.Lock()
		tunnel := rc.TunnelStore.TunnelByCluster[remoteClusterID]
		rc.TunnelStore.Mu.Unlock()

		if tunnel == nil {
			http.Error(w, "Tunnel not found", http.StatusInternalServerError)
			return nil
		}

		tunnel.Mu.Lock()
		isConnected := tunnel.WsConn != nil
		tunnel.Mu.Unlock()
		if !isConnected {
			http.Error(w, "Tunnel not connected", http.StatusInternalServerError)
			return nil
		}

		prefix := fmt.Sprintf("/1.0/remote-cluster/%s/tunnel", url.PathEscape(remoteClusterName))
		path := strings.TrimPrefix(r.URL.Path, prefix)
		req := types.ClusterManagerTunnelRequest{
			ID:      id,
			Method:  r.Method,
			Path:    path,
			Headers: extractHeaders(r),
			Body:    body,
		}

		tunnel.Mu.Lock()
		err = tunnel.WsConn.WriteJSON(req)
		tunnel.Mu.Unlock()

		if err != nil {
			logger.Log.Errorw("Failed to send request over WebSocket", "error", err)
			http.Error(w, "Failed to send request", http.StatusInternalServerError)
			ensureClosed(tunnel)
			return nil
		}

		// Wait for response
		ch := make(chan types.ClusterManagerTunnelResponse)
		tunnel.PendingCalls[id] = ch
		defer delete(tunnel.PendingCalls, id)

		select {
		case resp := <-ch:
			writeResponse(w, resp)
		case <-time.After(5 * time.Second):
			http.Error(w, "Timeout", http.StatusGatewayTimeout)
		}

		return nil
	}
}

func extractHeaders(r *http.Request) map[string]string {
	headers := make(map[string]string)
	for k, v := range r.Header {
		if len(v) > 0 {
			headers[k] = v[0]
		}
	}
	return headers
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
	// Write headers
	for k, v := range resp.Headers {
		value := v
		if k == "Location" {
			// Rewrite Location header, as the LXD-UI redirects and we need to ensure it points to the nested path
			value = strings.ReplaceAll(value, "/ui", "/1.0/remote-cluster/lxd-ui/ui")
		}
		w.Header().Set(k, value)
	}

	w.WriteHeader(resp.Status)
	_, err := w.Write(resp.Body)
	if err != nil {
		logger.Log.Error("error writing response body: %v", err)
	}
}

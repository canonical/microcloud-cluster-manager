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
	"strings"
	"time"

	"github.com/canonical/lxd/lxd/request"
	"github.com/canonical/lxd/lxd/response"
	"github.com/canonical/microcloud-cluster-manager/internal/app/cluster-connector/core/auth"
	"github.com/canonical/microcloud-cluster-manager/internal/app/cluster-connector/core/certificate"
	"github.com/canonical/microcloud-cluster-manager/internal/app/cluster-connector/core/rate_limit"
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

func remoteClusterTunnel(rc types.RouteConfig) types.EndpointHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		remoteClusterName, err := url.PathUnescape(mux.Vars(r)["remoteClusterName"])
		if err != nil {
			return response.SmartError(err).Render(w, r)
		}

		authorizationHeader := r.Header.Get("Authorization")
		if authorizationHeader == "" {
			return response.SmartError(errors.New("Authorization header is missing")).Render(w, r)
		}

		userSecret := r.Header.Get("X-User-Secret")
		if userSecret == "" {
			return response.SmartError(errors.New("X-User-Secret header is missing")).Render(w, r)
		}

		var remoteClusterID int
		err = rc.DB.Transaction(r.Context(), func(ctx context.Context, tx *sqlx.Tx) error {
			remoteClusterID, err = store.GetRemoteClusterID(ctx, tx, remoteClusterName)
			if err != nil {
				return err
			}

			return nil
		})
		if err != nil {
			return response.SmartError(err).Render(w, r)
		}

		id, err := uuid.NewV7()
		if err != nil {
			logger.Log.Errorw("Failed to generate request ID", "error", err)
			return response.SmartError(errors.New("Failed to generate request ID")).Render(w, r)
		}

		body, err := readBody(r)
		if err != nil {
			logger.Log.Errorw("Failed to read request body", "error", err)
			return response.SmartError(errors.New("Failed to read request body")).Render(w, r)
		}

		rc.TunnelStore.Mu.Lock()
		tunnel := rc.TunnelStore.TunnelByCluster[remoteClusterID]
		rc.TunnelStore.Mu.Unlock()

		if tunnel == nil {
			return response.SmartError(errors.New("Tunnel not found")).Render(w, r)
		}

		headers := http.Header{}
		headers.Set("Authorization", authorizationHeader)
		session, err := getUserSession(tunnel, authorizationHeader, userSecret)
		if err != nil {
			return response.SmartError(errors.New("Error loading LXD session value")).Render(w, r)
		}

		if session != "" {
			headers.Set("Cookie", "session="+session)
		}

		prefix := fmt.Sprintf("/%s/remote-cluster/%s/tunnel", rc.Env.APIVersion, url.PathEscape(remoteClusterName))
		path := strings.TrimPrefix(r.URL.Path, prefix)
		req := types.ClusterManagerTunnelRequest{
			UUID:    id.String(),
			Method:  r.Method,
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
			return response.SmartError(errors.New("Tunnel not connected")).Render(w, r)
		}
		err = wsConn.WriteJSON(req)
		tunnel.Mu.Unlock()

		if err != nil {
			logger.Log.Errorw("Failed to send request over WebSocket", "error", err)
			ensureClosed(tunnel)
			return response.SmartError(errors.New("Failed to send request")).Render(w, r)
		}

		// Wait for response
		select {
		case resp := <-ch:
			setUserSession(resp, tunnel, authorizationHeader, userSecret)
			writeResponse(w, resp)
		case <-time.After(15 * time.Second):
			return response.SmartError(errors.New("Timeout")).Render(w, r)
		case <-r.Context().Done():
			return response.SmartError(errors.New("Client disconnected")).Render(w, r)
		}

		return nil
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

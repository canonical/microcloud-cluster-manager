package v1

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/canonical/lxd/lxd/request"
	"github.com/canonical/lxd/lxd/response"
	"github.com/canonical/microcloud-cluster-manager/internal/app/cluster-connector/core/auth"
	"github.com/canonical/microcloud-cluster-manager/internal/app/cluster-connector/core/certificate"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/api/models/v1"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/database/store"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/logger"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/types"
	"github.com/golang/snappy"
	"github.com/jmoiron/sqlx"
	"github.com/prometheus/common/expfmt"
	"github.com/prometheus/prometheus/prompb"
)

// RemoteCluster is the API endpoint for managing remote clusters.
var RemoteCluster = types.RouteGroup{
	Prefix: "remote-cluster",
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
		auth.AuthMiddleware,
	},
	Endpoints: []types.Endpoint{
		{
			Path:    "status",
			Method:  http.MethodPost,
			Handler: remoteClusterStatusPost,
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
			return response.SmartError(err).Render(w, r)
		}

		// check if tokenFromDb has expired
		if time.Now().After(tokenFromDb.Expiry) {
			return response.Forbidden(fmt.Errorf("tokenFromDb has expired")).Render(w, r)
		}

		isTokenValid := strings.EqualFold(payload.Token, tokenFromDb.EncodedToken)
		if !isTokenValid {
			return response.Forbidden(err).Render(w, r)
		}

		// Create remote cluster entry and delete tokenFromDb in a single db transaction
		var remoteClusterID int
		err = rc.DB.Transaction(r.Context(), func(ctx context.Context, tx *sqlx.Tx) error {
			// create remote cluster entry
			newRemoteCluster, err := store.CreateRemoteCluster(ctx, tx, store.RemoteCluster{
				Name:               payload.ClusterName,
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
				RemoteClusterID:  newRemoteCluster.ID,
				MemberStatuses:   json.RawMessage("[]"),
				InstanceStatuses: json.RawMessage("[]"),
			})

			if err != nil {
				return err
			}

			// delete remote cluster tokenFromDb
			return store.DeleteRemoteClusterToken(ctx, tx, payload.ClusterName)
		})

		if err != nil {
			return response.SmartError(err).Render(w, r)
		}

		verifier, ok := rc.Auth.(*auth.MtlsAuthenticator)
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
			return response.BadRequest(err).Render(w, r)
		}

		remoteClusterID, err := request.GetCtxValue[int](r.Context(), auth.CtxRemoteClusterID)
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
				Status:             string(models.ACTIVE),
				JoinedAt:           time.Now(),
				ClusterCertificate: dbRemoteCluster.ClusterCertificate,
			}

			err = store.UpdateRemoteCluster(ctx, tx, dbRemoteCluster.Name, newCluster)
			if err != nil {
				return err
			}

			if payload.Metrics == "" {
				return nil
			}

			if rc.Env.PrometheusBaseURL == "" || rc.Env.PrometheusBaseURL == "http://base" {
				logger.Log.Infow("Prometheus base URL is not configured, metrics received by cluster are discarded.", "remote cluster", remoteClusterID)
				return nil
			}

			timeSeries, err := parsePrometheusMetrics(payload.Metrics, dbRemoteCluster.Name)
			if err != nil {
				logger.Log.Warnw("Failed to parse metrics", "remote cluster", remoteClusterID, "err", err)
				return fmt.Errorf("failed to parse Prometheus metrics: %w", err)
			}

			err = forwardMetricsToPrometheus(timeSeries, rc)
			if err != nil {
				logger.Log.Warnw("Failed to forward metrics to Prometheus", "remote cluster", remoteClusterID, "err", err)
				return fmt.Errorf("failed to forward metrics to Prometheus: %w", err)
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
			ClusterManagerAddress: rc.Env.ClusterConnectorAddress,
		}).Render(w, r)
	}
}

// Parse the incoming Prometheus metrics (text format).
func parsePrometheusMetrics(metricsText string, jobName string) ([]prompb.TimeSeries, error) {
	var parser expfmt.TextParser
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
	defer resp.Body.Close()

	if resp.StatusCode > 299 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to send metrics to Prometheus, status: %s, response: %s", resp.Status, string(body))
	}

	return nil
}

func remoteClusterDelete(rc types.RouteConfig) types.EndpointHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		remoteClusterID, err := request.GetCtxValue[int](r.Context(), auth.CtxRemoteClusterID)
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

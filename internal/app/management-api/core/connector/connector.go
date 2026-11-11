// Package connector is the client of the management API for the cluster connector.
package connector

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/canonical/lxd/shared"
	"github.com/canonical/lxd/shared/api"
	"github.com/canonical/microcloud-cluster-manager/internal/app/management-api/core/auth"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/logger"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/types"
)

// maxRequestBytes limits the request body that is buffered to forward it to the cluster connector.
const maxRequestBytes = 1 << 20 // 1 MiB

// Forward sends the request to the cluster connector member holding the tunnel of the remote
// cluster and returns its response, which the caller must close. The member URL is the tunnel manager member URL of
// the remote cluster, which the caller has to provide.
func Forward(rc types.RouteConfig, r *http.Request, remoteClusterName string, clusterConnectorPath string, memberURL string) (*http.Response, error) {
	if memberURL == "" {
		return nil, api.StatusErrorf(http.StatusBadRequest, "remote cluster does not have an active tunnel")
	}

	reqURL := memberURL + "/1.0/remote-cluster/" + url.PathEscape(remoteClusterName) + clusterConnectorPath

	sessionProvider, ok := r.Context().Value(auth.SessionCredentialsProviderKey).(*auth.SessionCredentialsProvider)
	if !ok {
		return nil, fmt.Errorf("failed to get session credentials provider from the request context")
	}

	accessToken, err := sessionProvider.GetAccessToken()
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	userSecret, err := sessionProvider.GetUserSecret()
	if err != nil {
		return nil, fmt.Errorf("failed to get user secret: %w", err)
	}

	// Buffer the body so that it can be sent again if the request needs to be retried after a token refresh.
	body, err := io.ReadAll(http.MaxBytesReader(nil, r.Body, maxRequestBytes))
	if err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			return nil, api.StatusErrorf(http.StatusRequestEntityTooLarge, "request body is larger than %d bytes", maxRequestBytes)
		}

		return nil, fmt.Errorf("failed to read request body: %w", err)
	}

	defer func() {
		err := r.Body.Close()
		if err != nil {
			logger.Log.Warnw("Failed to close request body", "error", err)
		}
	}()

	resp, err := sendInternalRequest(r, accessToken, userSecret, reqURL, rc, body)
	if err != nil {
		return nil, err
	}

	// Only replay requests that have no side effects. A 401/403 from a non-idempotent request may be the
	// result of a partially applied change, so replaying it could apply the change twice.
	isRetriable := r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodOptions
	needsTokenRefresh := isRetriable && (resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden)
	if needsTokenRefresh {
		refreshCtx, cancelRefresh := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancelRefresh()
		accessToken, err = sessionProvider.RefreshSessionTokens(refreshCtx)
		if err != nil {
			_ = resp.Body.Close()
			return nil, fmt.Errorf("failed to refresh access token: %w", err)
		}

		err = resp.Body.Close()
		if err != nil {
			logger.Log.Warnw("Failed to close response body before retry", "error", err)
		}

		resp, err = sendInternalRequest(r, accessToken, userSecret, reqURL, rc, body)
		if err != nil {
			return nil, fmt.Errorf("failed to send request after token refresh: %w", err)
		}
	}
	return resp, nil
}

// sendInternalRequest sends the request to the cluster connector over mTLS.
func sendInternalRequest(r *http.Request, accessToken string, userSecret string, reqURL string, rc types.RouteConfig, body []byte) (*http.Response, error) {
	req, err := http.NewRequestWithContext(r.Context(), r.Method, reqURL, bytes.NewReader(body))
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
			// The transport is not reused, so close the connection with the response instead of leaking it.
			DisableKeepAlives: true,
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

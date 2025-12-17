package helpers

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"time"

	"github.com/canonical/lxd/shared"
	"github.com/canonical/lxd/shared/api"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/api/models/v1"
)

// FindRemoteCluster search for a remote cluster by name.
func FindRemoteCluster(env *Environment, remoteClusterName string) (*models.RemoteCluster, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	headers, err := env.ManagementAPILoginHeaders()
	if err != nil {
		return nil, err
	}

	certPublicKey, err := env.ManagementAPICert().PublicKeyX509()
	if err != nil {
		return nil, err
	}

	tlsClient, err := NewTLSHTTPClient(api.URL{}, nil, certPublicKey, env.ManagementAPIHost())
	if err != nil {
		return nil, err
	}

	output := &models.RemoteCluster{}
	path := api.NewURL().Scheme("https").Host(env.ManagementAPIHostPort()).Path("1.0", "remote-cluster", remoteClusterName)
	err = tlsClient.Query(ctx, http.MethodGet, path, nil, output, headers)
	if err != nil {
		return nil, err
	}

	return output, nil
}

// DeleteRemoteCluster deletes a remote cluster by name.
func DeleteRemoteCluster(env *Environment, remoteClusterName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	certPublicKey, err := env.ManagementAPICert().PublicKeyX509()
	if err != nil {
		return err
	}

	tlsClient, err := NewTLSHTTPClient(api.URL{}, nil, certPublicKey, env.ManagementAPIHost())
	if err != nil {
		return err
	}

	path := api.NewURL().Scheme("https").Host(env.ManagementAPIHostPort()).Path("1.0", "remote-cluster", remoteClusterName)
	return tlsClient.Query(ctx, http.MethodDelete, path, nil, nil, nil)
}

func GetRandomName(baseName string) string {
	randomNumber := rand.Intn(10000)
	return fmt.Sprintf("%s-%d", baseName, randomNumber)
}

// SendJoinRequest sets up a tls http client and sends a join request to Cluster Manager.
func SendJoinRequest(env *Environment, tokenData models.RemoteClusterTokenBody) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// generate dedicated lxd cert for Cluster Manager
	clientCert, err := shared.KeyPairAndCA(env.CertDir(), tokenData.ServerName, shared.CertClient, shared.CertOptions{AddHosts: false})
	if err != nil {
		return err
	}

	clusterConnectorCert := env.ClusterConnectorCert()
	clusterConnectorCertPublicKey, err := clusterConnectorCert.PublicKeyX509()
	if err != nil {
		return err
	}

	tlsClient, err := NewTLSHTTPClient(api.URL{}, clientCert, clusterConnectorCertPublicKey, env.ClusterConnectorHost())
	if err != nil {
		return err
	}

	encodedToken, err := tokenData.Encode()
	if err != nil {
		return err
	}

	input := models.RemoteClusterPost{
		ClusterName:        tokenData.ServerName,
		ClusterCertificate: string(clientCert.PublicKey()),
		Token:              encodedToken,
	}

	path := api.NewURL().Scheme("https").Host(tokenData.Addresses[0]).Path("1.0", "remote-cluster")

	return tlsClient.Query(ctx, http.MethodPost, path, input, nil, nil)
}

// RegisterRemoteCluster creates a remote cluster join token and sends a join request to Cluster Manager.
func RegisterRemoteCluster(env *Environment, remoteClusterName string) (*models.RemoteClusterTokenBody, error) {
	tokenData, err := CreateAndReturnRemoteClusterJoinToken(env, remoteClusterName, time.Time{})
	if err != nil {
		return nil, err
	}

	err = SendJoinRequest(env, tokenData)
	if err != nil {
		return nil, err
	}
	return &tokenData, nil
}

// SendStatusUpdate sends a status update to the Cluster Manager.
func SendStatusUpdate(env *Environment, tokenData models.RemoteClusterTokenBody, statusData models.RemoteClusterStatusPost) (*models.RemoteClusterStatusPostResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	clientCert, err := shared.KeyPairAndCA(env.CertDir(), tokenData.ServerName, shared.CertClient, shared.CertOptions{AddHosts: false})
	if err != nil {
		return nil, err
	}

	clusterConnectorCert := env.ClusterConnectorCert()
	clusterConnectorCertPublicKey, err := clusterConnectorCert.PublicKeyX509()
	if err != nil {
		return nil, err
	}

	tlsClient, err := NewTLSHTTPClient(api.URL{}, clientCert, clusterConnectorCertPublicKey, env.ClusterConnectorHost())
	if err != nil {
		return nil, err
	}

	var output models.RemoteClusterStatusPostResponse
	path := api.NewURL().Scheme("https").Host(tokenData.Addresses[0]).Path("1.0", "remote-cluster", "status")
	err = tlsClient.Query(ctx, http.MethodPost, path, statusData, &output, nil)
	return &output, err
}

func CreateStatusPostData() models.RemoteClusterStatusPost {
	return models.RemoteClusterStatusPost{
		CPUTotalCount:     8,
		CPULoad1:          "0.1",
		CPULoad5:          "0.2",
		CPULoad15:         "0.3",
		MemoryTotalAmount: 1024,
		MemoryUsage:       512,
		InstanceStatuses: []models.StatusDistribution{
			{Status: "running", Count: 1},
			{Status: "stopped", Count: 2},
		},
		MemberStatuses: []models.StatusDistribution{
			{Status: "active", Count: 1},
			{Status: "inactive", Count: 2},
		},
		StoragePoolUsages: []models.StoragePoolUsage{
			{Name: "default", Total: 1024, Usage: 512},
			{Name: "data", Total: 2048, Usage: 1024},
		},
	}
}

// QueryPrometheus sends a query to the Prometheus API and returns the response as a string.
func QueryPrometheus(env *Environment, query string) (string, error) {
	config, err := getConfiguration(env)
	if err != nil {
		return "", err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	prometheusBaseURL, err := url.Parse(config.PrometheusBaseURL.Value)

	if err != nil {
		return "", err
	}

	url := api.NewURL().Scheme(prometheusBaseURL.Scheme).Host(prometheusBaseURL.Host).Path("api", "v1", "query")
	q := url.Query()
	q.Set("query", query)
	url.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
	if err != nil {
		return "", err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return "", err
	}
	result := string(body)
	return result, nil
}

func getConfiguration(env *Environment) (*models.Configuration, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	headers, err := env.ManagementAPILoginHeaders()
	if err != nil {
		return nil, err
	}

	certPublicKey, err := env.ManagementAPICert().PublicKeyX509()
	if err != nil {
		return nil, err
	}

	tlsClient, err := NewTLSHTTPClient(api.URL{}, nil, certPublicKey, env.ManagementAPIHost())
	if err != nil {
		return nil, err
	}
	output := &models.Configuration{}
	path := api.NewURL().Scheme("https").Host(env.ManagementAPIHostPort()).Path("1.0", "configuration")
	err = tlsClient.Query(ctx, http.MethodGet, path, nil, output, headers)
	if err != nil {
		return nil, err
	}

	return output, nil
}

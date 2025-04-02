package main

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/canonical/lxd-cluster-manager/internal/pkg/api/models/v1"
	"github.com/canonical/lxd-cluster-manager/test/helpers"
	"github.com/canonical/lxd/shared"
	"github.com/canonical/lxd/shared/api"
)

func testRemoteClusterSuccess(env *helpers.Environment) (testName string, testFunc func(t *testing.T)) {
	return "lxd remote cluster join and status updates under normal conditions", func(t *testing.T) {
		remoteClusterName := "remote_cluster_e2e"
		var condition string
		var err error
		var tokenData models.RemoteClusterTokenBody

		{
			condition = "Should be able to create token with valid data"
			tokenData, err = helpers.CreateAndReturnRemoteClusterJoinToken(env, remoteClusterName, time.Time{})
			if err != nil {
				helpers.LogTestOutcome(t, condition, err)
			}

			if tokenData.ServerName == "" {
				err = fmt.Errorf("invalid server_name")
				helpers.LogTestOutcome(t, condition, err)
			}

			if tokenData.Fingerprint == "" {
				err = fmt.Errorf("invalid fingerprint")
				helpers.LogTestOutcome(t, condition, err)
			}

			if tokenData.Addresses[0] != env.ClusterConnectorHostPort() {
				err = fmt.Errorf("invalid address")
				helpers.LogTestOutcome(t, condition, err)
			}

			if tokenData.Secret == "" {
				err = fmt.Errorf("invalid secret")
				helpers.LogTestOutcome(t, condition, err)
			}

			if time.Time.Equal(tokenData.ExpiresAt, time.Time{}) {
				err = fmt.Errorf("invalid expiry")
				helpers.LogTestOutcome(t, condition, err)
			}

			helpers.LogTestOutcome(t, condition, nil)
		}

		{
			condition = "Should be able to receive a join request"
			err = sendJoinRequest(env, tokenData)
			helpers.LogTestOutcome(t, condition, err)
		}

		{
			condition = "Should be able to get remote cluster with ACTIVE status"
			remoteCluster, err := helpers.FindRemoteCluster(env, remoteClusterName)
			if err != nil {
				helpers.LogTestOutcome(t, condition, err)
			}

			if remoteCluster.Status != string(models.ACTIVE) {
				err = fmt.Errorf("invalid remote cluster status")
				helpers.LogTestOutcome(t, condition, err)
			}

			helpers.LogTestOutcome(t, condition, nil)
		}

		{
			condition = "Should have deleted the remote cluster join token after receiving join request"
			token, err := helpers.FindToken(env, remoteClusterName)
			if err != nil {
				helpers.LogTestOutcome(t, condition, err)
			}

			if token != (models.RemoteClusterToken{}) {
				err = fmt.Errorf("token not deleted")
				helpers.LogTestOutcome(t, condition, err)
			}

			helpers.LogTestOutcome(t, condition, nil)
		}

		{
			condition = "Should be able to receive a status update"
			response, err := sendStatusUpdate(env, tokenData)
			if err != nil {
				helpers.LogTestOutcome(t, condition, err)
			}

			expected := env.ClusterConnectorHostPort()

			if !reflect.DeepEqual(response.ClusterManagerAddress, expected) {
				fmt.Println(response.ClusterManagerAddress)
				fmt.Println(expected)
				err = fmt.Errorf("invalid Cluster Manager address")
				helpers.LogTestOutcome(t, condition, err)
			}

			helpers.LogTestOutcome(t, condition, nil)
		}

		{
			condition = "Should be able to get remote cluster status"
			remoteCluster, err := helpers.FindRemoteCluster(env, remoteClusterName)
			if err != nil {
				helpers.LogTestOutcome(t, condition, err)
			}

			if remoteCluster.CPUTotalCount != 8 {
				err = fmt.Errorf("invalid CPU total count")
				helpers.LogTestOutcome(t, condition, err)
			}

			if remoteCluster.MemoryTotalAmount != 1024 {
				err = fmt.Errorf("invalid memory total amount")
				helpers.LogTestOutcome(t, condition, err)
			}

			if remoteCluster.DiskTotalSize != 1024 {
				err = fmt.Errorf("invalid disk total size")
				helpers.LogTestOutcome(t, condition, err)
			}

			if !reflect.DeepEqual(remoteCluster.InstanceStatuses, []models.StatusDistribution{
				{Status: "running", Count: 1},
				{Status: "stopped", Count: 2},
			}) {
				err = fmt.Errorf("invalid instance statuses")
				helpers.LogTestOutcome(t, condition, err)
			}

			if !reflect.DeepEqual(remoteCluster.MemberStatuses, []models.StatusDistribution{
				{Status: "active", Count: 1},
				{Status: "inactive", Count: 2},
			}) {
				err = fmt.Errorf("invalid member statuses")
				helpers.LogTestOutcome(t, condition, err)
			}

			helpers.LogTestOutcome(t, condition, nil)
		}

		env.RemoveRemoteClusterToken(remoteClusterName)
		env.RemoveRemoteCluster(remoteClusterName)
	}
}

// sendJoinRequest sets up a tls http client and sends a join request to Cluster Manager.
func sendJoinRequest(env *helpers.Environment, tokenData models.RemoteClusterTokenBody) error {
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

	tlsClient, err := helpers.NewTLSHTTPClient(api.URL{}, clientCert, clusterConnectorCertPublicKey, env.ClusterConnectorHost())
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

// sendStatusUpdate sends a status update to the Cluster Manager with the correct client certificate.
func sendStatusUpdate(env *helpers.Environment, tokenData models.RemoteClusterTokenBody) (*models.RemoteClusterStatusPostResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
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

	tlsClient, err := helpers.NewTLSHTTPClient(api.URL{}, clientCert, clusterConnectorCertPublicKey, env.ClusterConnectorHost())
	if err != nil {
		return nil, err
	}

	input := models.RemoteClusterStatusPost{
		CPUTotalCount:     8,
		CPULoad1:          "0.1",
		CPULoad5:          "0.2",
		CPULoad15:         "0.3",
		MemoryTotalAmount: 1024,
		MemoryUsage:       512,
		DiskTotalSize:     1024,
		DiskUsage:         512,
		InstanceStatuses: []models.StatusDistribution{
			{Status: "running", Count: 1},
			{Status: "stopped", Count: 2},
		},
		MemberStatuses: []models.StatusDistribution{
			{Status: "active", Count: 1},
			{Status: "inactive", Count: 2},
		},
	}

	var output models.RemoteClusterStatusPostResponse
	path := api.NewURL().Scheme("https").Host(tokenData.Addresses[0]).Path("1.0", "remote-cluster", "status")
	err = tlsClient.Query(ctx, http.MethodPost, path, input, &output, nil)

	return &output, err
}

package main

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/canonical/lxd/shared"
	"github.com/canonical/lxd/shared/api"

	"github.com/canonical/lxd-cluster-manager/internal/api/types"
	"github.com/canonical/lxd-cluster-manager/test/helpers"
)

func testRemoteClusterSuccess(env *helpers.Environment) (testName string, testFunc func(t *testing.T)) {
	return "lxd remote cluster join and status updates under normal conditions", func(t *testing.T) {
		remoteClusterName := "remote_cluster_control_e2e"
		var condition string
		var err error
		var tokenData types.RemoteClusterTokenBody

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

			if tokenData.Addresses[0] != "0.0.0.0:9110" {
				err = fmt.Errorf("invalid address")
				helpers.LogTestOutcome(t, condition, err)
			}

			if tokenData.Secret == "" {
				err = fmt.Errorf("invalid secret")
				helpers.LogTestOutcome(t, condition, err)
			}

			if tokenData.ExpiresAt == (time.Time{}) {
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
			condition = "Should be able to get remote cluster with PENDING_APPROVAL status"
			remoteCluster, err := helpers.FindRemoteCluster(env, remoteClusterName)
			if err != nil {
				helpers.LogTestOutcome(t, condition, err)
			}

			if remoteCluster.Status != string(types.PENDING_APPROVAL) {
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

			if token != (types.RemoteClusterToken{}) {
				err = fmt.Errorf("token not deleted")
				helpers.LogTestOutcome(t, condition, err)
			}

			helpers.LogTestOutcome(t, condition, nil)
		}

		{
			condition = "Should be able to approve a join request"
			err = approveJoinRequest(env, remoteClusterName)
			if err != nil {
				helpers.LogTestOutcome(t, condition, err)
			}

			remoteCluster, err := helpers.FindRemoteCluster(env, remoteClusterName)
			if err != nil {
				helpers.LogTestOutcome(t, condition, err)
			}

			if remoteCluster.Status != string(types.ACTIVE) {
				err = fmt.Errorf("invalid remote cluster status")
				helpers.LogTestOutcome(t, condition, err)
			}

			helpers.LogTestOutcome(t, condition, nil)
		}

		{
			condition = "Should be able to receive a status update"
			actual, err := sendStatusUpdate(env, tokenData)
			if err != nil {
				helpers.LogTestOutcome(t, condition, err)
			}

			expected := types.RemoteClusterStatusPostResponse{
				ClusterManagerAddresses: []string{"0.0.0.0:9110"},
			}

			if !reflect.DeepEqual(*actual, expected) {
				err = fmt.Errorf("invalid Cluster Manager addresses")
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

			if !reflect.DeepEqual(remoteCluster.InstanceStatuses, []types.StatusDistribution{
				{Status: "running", Count: 1},
				{Status: "stopped", Count: 2},
			}) {
				err = fmt.Errorf("invalid instance statuses")
				helpers.LogTestOutcome(t, condition, err)
			}

			if !reflect.DeepEqual(remoteCluster.MemberStatuses, []types.StatusDistribution{
				{Status: "active", Count: 1},
				{Status: "inactive", Count: 2},
			}) {
				err = fmt.Errorf("invalid member statuses")
				helpers.LogTestOutcome(t, condition, err)
			}

			helpers.LogTestOutcome(t, condition, nil)
		}
	}
}

// sendJoinRequest sets up a tls http client and sends a join request to Cluster Manager.
func sendJoinRequest(env *helpers.Environment, tokenData types.RemoteClusterTokenBody) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// generate dedicated lxd cert for Cluster Manager
	clientCert, err := shared.KeyPairAndCA(env.CertDir(), tokenData.ServerName, shared.CertClient, false)
	if err != nil {
		return err
	}

	clusterCert, err := env.GetClusterCert()
	if err != nil {
		return err
	}

	clusterCertPublicKey, err := clusterCert.PublicKeyX509()
	if err != nil {
		return err
	}

	tlsClient, err := helpers.NewTLSHTTPClient(api.URL{}, clientCert, clusterCertPublicKey)
	if err != nil {
		return err
	}

	input := struct {
		ClusterName             string `json:"cluster_name"`
		RemotClusterCertificate string `json:"cluster_certificate"`
	}{
		ClusterName:             tokenData.ServerName,
		RemotClusterCertificate: string(clientCert.PublicKey()),
	}

	path := api.NewURL().Scheme("https").Host(tokenData.Addresses[0]).Path("1.0", "remote-clusters")
	adjustHeaders := func(req *http.Request) error {
		mac := hmac.New(sha256.New, []byte(tokenData.Secret))
		inputBytes, err := json.Marshal(input)
		if err != nil {
			return err
		}

		mac.Write(inputBytes)
		req.Header.Set("X-CLUSTER-SIGNATURE", base64.StdEncoding.EncodeToString(mac.Sum(nil)))
		return nil
	}

	return tlsClient.Query(ctx, http.MethodPost, path, input, nil, adjustHeaders)
}

// approveJoinRequest approves a join request for a remote cluster.
func approveJoinRequest(env *helpers.Environment, remoteClusterName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	unixClient, err := helpers.NewUnixHTTPClient(env.ControlSocketURL())
	if err != nil {
		return err
	}

	input := types.RemoteClusterPatch{
		Status: types.ACTIVE,
	}

	path := api.NewURL().Path("1.0", "remote-clusters", remoteClusterName)
	return unixClient.Query(ctx, http.MethodPatch, path, input, nil, nil)
}

// sendStatusUpdate sends a status update to the Cluster Manager with the correct client certificate.
func sendStatusUpdate(env *helpers.Environment, tokenData types.RemoteClusterTokenBody) (*types.RemoteClusterStatusPostResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	clientCert, err := shared.KeyPairAndCA(env.CertDir(), tokenData.ServerName, shared.CertClient, false)
	if err != nil {
		return nil, err
	}

	clusterCert, err := env.GetClusterCert()
	if err != nil {
		return nil, err
	}

	clusterCertPublicKey, err := clusterCert.PublicKeyX509()
	if err != nil {
		return nil, err
	}

	tlsClient, err := helpers.NewTLSHTTPClient(api.URL{}, clientCert, clusterCertPublicKey)
	if err != nil {
		return nil, err
	}

	input := types.RemoteClusterStatusPost{
		CPUTotalCount:     8,
		CPULoad1:          "0.1",
		CPULoad5:          "0.2",
		CPULoad15:         "0.3",
		MemoryTotalAmount: 1024,
		MemoryUsage:       512,
		DiskTotalSize:     1024,
		DiskUsage:         512,
		InstanceStatuses: []types.StatusDistribution{
			{Status: "running", Count: 1},
			{Status: "stopped", Count: 2},
		},
		MemberStatuses: []types.StatusDistribution{
			{Status: "active", Count: 1},
			{Status: "inactive", Count: 2},
		},
	}

	var output types.RemoteClusterStatusPostResponse
	path := api.NewURL().Scheme("https").Host(tokenData.Addresses[0]).Path("1.0", "remote-clusters", "status")
	err = tlsClient.Query(ctx, http.MethodPost, path, input, &output, nil)

	return &output, err
}

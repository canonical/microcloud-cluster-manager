package main

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/canonical/lxd/shared/api"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/api/models/v1"
	"github.com/canonical/microcloud-cluster-manager/test/helpers"
)

func testRemoteClusterStatusNoCert(env *helpers.Environment) (testName string, testFunc func(t *testing.T)) {
	return "lxd remote cluster status update with no certificate", func(t *testing.T) {
		remoteClusterName := helpers.GetRandomName("remote_cluster_status_no_cert")
		var condition string

		{
			condition = "Should fail status update request with no client certificate"

			tokenData, err := helpers.RegisterRemoteCluster(env, remoteClusterName)
			if err != nil {
				helpers.LogTestOutcome(t, condition, err)
			}

			err = sendStatusUpdateNoCert(env, *tokenData)
			if err != nil && err.Error() == "Forbidden" {
				err = nil
			} else {
				err = errors.New("expected forbidden error not received")
			}

			helpers.LogTestOutcome(t, condition, err)
		}

		env.RemoveRemoteClusterToken(remoteClusterName)
		env.RemoveRemoteCluster(remoteClusterName)
	}
}

func testRemoteClusterStatusInvalidCert(env *helpers.Environment) (testName string, testFunc func(t *testing.T)) {
	return "lxd remote cluster status update with an invalid certificate", func(t *testing.T) {
		remoteClusterName := helpers.GetRandomName("remote_cluster_status_invalid_cert")
		var condition string

		{
			condition = "Should fail status update request with an invalid certificate"

			tokenData, err := helpers.RegisterRemoteCluster(env, remoteClusterName)
			if err != nil {
				helpers.LogTestOutcome(t, condition, err)
			}

			err = sendStatusUpdateInvalidCert(env, *tokenData)
			if err != nil && err.Error() == "Forbidden" {
				err = nil
			} else {
				err = errors.New("expected forbidden error not received")
			}

			helpers.LogTestOutcome(t, condition, err)
		}

		env.RemoveRemoteClusterToken(remoteClusterName)
		env.RemoveRemoteCluster(remoteClusterName)
	}
}

// sendStatusUpdateNoCert sends a status update to the Cluster Manager with no client certificate.
func sendStatusUpdateNoCert(env *helpers.Environment, tokenData models.RemoteClusterTokenBody) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	clusterConnectorCert := env.ClusterConnectorCert()
	clusterConnectorCertPublicKey, err := clusterConnectorCert.PublicKeyX509()
	if err != nil {
		return err
	}

	tlsClient, err := helpers.NewTLSHTTPClient(api.URL{}, nil, clusterConnectorCertPublicKey, env.ClusterConnectorHost())
	if err != nil {
		return err
	}

	path := api.NewURL().Scheme("https").Host(tokenData.Addresses[0]).Path("1.0", "remote-cluster", "status")
	_, err = tlsClient.Query(ctx, http.MethodPost, path, nil, nil, nil)
	return err
}

// sendStatusUpdateInvalidCert sends a status update to the Cluster Manager with a client certificate that was not sent with the join request.
func sendStatusUpdateInvalidCert(env *helpers.Environment, tokenData models.RemoteClusterTokenBody) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	clusterConnectorCert := env.ClusterConnectorCert()
	clusterConnectorCertPublicKey, err := clusterConnectorCert.PublicKeyX509()
	if err != nil {
		return err
	}

	// send cluster cert as client cert, this should cause cluster manager to not find the remote clsuter
	tlsClient, err := helpers.NewTLSHTTPClient(api.URL{}, clusterConnectorCert, clusterConnectorCertPublicKey, env.ClusterConnectorHost())
	if err != nil {
		return err
	}

	path := api.NewURL().Scheme("https").Host(tokenData.Addresses[0]).Path("1.0", "remote-cluster", "status")
	_, err = tlsClient.Query(ctx, http.MethodPost, path, nil, nil, nil)
	return err
}

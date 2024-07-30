package main

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/canonical/lxd/shared/api"

	"github.com/canonical/lxd-cluster-manager/internal/api/types"
	"github.com/canonical/lxd-cluster-manager/test/helpers"
)

func testRemoteClusterStatusNoCert(env *helpers.Environment) (testName string, testFunc func(t *testing.T)) {
	return "lxd remote cluster status update with no certificate", func(t *testing.T) {
		remoteClusterName := "remote_cluster_status_no_cert"
		var condition string

		{
			condition = "Should fail status update request with no client certificate"

			tokenData, err := helpers.CreateAndReturnRemoteClusterJoinToken(env, remoteClusterName, time.Time{})
			if err != nil {
				helpers.LogTestOutcome(t, condition, err)
			}

			err = sendJoinRequest(env, tokenData)
			if err != nil {
				helpers.LogTestOutcome(t, condition, err)
			}

			err = approveJoinRequest(env, remoteClusterName)
			if err != nil {
				helpers.LogTestOutcome(t, condition, err)
			}

			err = sendStatusUpdateNoCert(env, tokenData)
			if err != nil && err.Error() == "tls is required" {
				err = nil
			}

			helpers.LogTestOutcome(t, condition, err)
		}
	}
}

func testRemoteClusterStatusInactive(env *helpers.Environment) (testName string, testFunc func(t *testing.T)) {
	return "lxd remote cluster status update with an inactive remote cluster", func(t *testing.T) {
		remoteClusterName := "remote_cluster_status_inactive"
		var condition string

		{
			condition = "Should fail status update request with an inactive remote cluster"

			tokenData, err := helpers.CreateAndReturnRemoteClusterJoinToken(env, remoteClusterName, time.Time{})
			if err != nil {
				helpers.LogTestOutcome(t, condition, err)
			}

			err = sendJoinRequest(env, tokenData)
			if err != nil {
				helpers.LogTestOutcome(t, condition, err)
			}

			err = approveJoinRequest(env, remoteClusterName)
			if err != nil {
				helpers.LogTestOutcome(t, condition, err)
			}

			_, err = sendStatusUpdate(env, tokenData)
			if err != nil && err.Error() == "remote cluster is pending approval" {
				err = nil
			}

			helpers.LogTestOutcome(t, condition, err)
		}
	}
}

func testRemoteClusterStatusInvalidCert(env *helpers.Environment) (testName string, testFunc func(t *testing.T)) {
	return "lxd remote cluster status update with an invalid certificate", func(t *testing.T) {
		remoteClusterName := "remote_cluster_status_invalid_cert"
		var condition string

		{
			condition = "Should fail status update request with an invalid certificate"

			tokenData, err := helpers.CreateAndReturnRemoteClusterJoinToken(env, remoteClusterName, time.Time{})
			if err != nil {
				helpers.LogTestOutcome(t, condition, err)
			}

			err = sendJoinRequest(env, tokenData)
			if err != nil {
				helpers.LogTestOutcome(t, condition, err)
			}

			err = approveJoinRequest(env, remoteClusterName)
			if err != nil {
				helpers.LogTestOutcome(t, condition, err)
			}

			err = sendStatusUpdateInvalidCert(env, tokenData)
			if err != nil && err.Error() == "invalid cluster certificate" {
				err = nil
			}

			helpers.LogTestOutcome(t, condition, err)
		}
	}
}

// sendStatusUpdateNoCert sends a status update to the Cluster Manager with no client certificate.
func sendStatusUpdateNoCert(env *helpers.Environment, tokenData types.RemoteClusterTokenBody) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	clusterCert, err := env.GetClusterCert()
	if err != nil {
		return err
	}

	clusterCertPublicKey, err := clusterCert.PublicKeyX509()
	if err != nil {
		return err
	}

	tlsClient, err := helpers.NewTLSHTTPClient(api.URL{}, nil, clusterCertPublicKey)
	if err != nil {
		return err
	}

	path := api.NewURL().Scheme("https").Host(tokenData.Addresses[0]).Path("1.0", "remote-clusters", "status")
	return tlsClient.Query(ctx, http.MethodPost, path, nil, nil, nil)
}

// sendStatusUpdateInvalidCert sends a status update to the Cluster Manager with a client certificate that was not sent with the join request.
func sendStatusUpdateInvalidCert(env *helpers.Environment, tokenData types.RemoteClusterTokenBody) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	clusterCert, err := env.GetClusterCert()
	if err != nil {
		return err
	}

	clusterCertPublicKey, err := clusterCert.PublicKeyX509()
	if err != nil {
		return err
	}

	// send cluster cert as client cert, this should cause clsuter manager to not find the remote clsuter
	tlsClient, err := helpers.NewTLSHTTPClient(api.URL{}, clusterCert, clusterCertPublicKey)
	if err != nil {
		return err
	}

	path := api.NewURL().Scheme("https").Host(tokenData.Addresses[0]).Path("1.0", "remote-clusters", "status")
	return tlsClient.Query(ctx, http.MethodPost, path, nil, nil, nil)
}

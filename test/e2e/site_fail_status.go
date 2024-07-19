package main

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/canonical/lxd/shared/api"

	"github.com/canonical/lxd-site-manager/internal/api/types"
	"github.com/canonical/lxd-site-manager/test/helpers"
)

func testSiteStatusNoCert(env *helpers.Environment) (testName string, testFunc func(t *testing.T)) {
	return "lxd site status update with no certificate", func(t *testing.T) {
		siteName := "site_status_no_cert"
		var condition string

		{
			condition = "Should fail status update request with no client certificate"

			tokenData, err := helpers.CreateAndReturnSiteJoinToken(env, siteName, time.Time{})
			if err != nil {
				helpers.LogTestOutcome(t, condition, err)
			}

			err = sendJoinRequest(env, tokenData)
			if err != nil {
				helpers.LogTestOutcome(t, condition, err)
			}

			err = approveJoinRequest(env, siteName)
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

func testSiteStatusInactiveSite(env *helpers.Environment) (testName string, testFunc func(t *testing.T)) {
	return "lxd site status update with an inactive site", func(t *testing.T) {
		siteName := "site_status_inactive"
		var condition string

		{
			condition = "Should fail status update request with an inactive site"

			tokenData, err := helpers.CreateAndReturnSiteJoinToken(env, siteName, time.Time{})
			if err != nil {
				helpers.LogTestOutcome(t, condition, err)
			}

			err = sendJoinRequest(env, tokenData)
			if err != nil {
				helpers.LogTestOutcome(t, condition, err)
			}

			err = approveJoinRequest(env, siteName)
			if err != nil {
				helpers.LogTestOutcome(t, condition, err)
			}

			err = sendStatusUpdate(env, tokenData)
			if err != nil && err.Error() == "site not found" {
				err = nil
			}

			helpers.LogTestOutcome(t, condition, err)
		}
	}
}

func testSiteStatusInvalidCert(env *helpers.Environment) (testName string, testFunc func(t *testing.T)) {
	return "lxd site status update with an invalid certificate", func(t *testing.T) {
		siteName := "site_status_invalid_cert"
		var condition string

		{
			condition = "Should fail status update request with an invalid certificate"

			tokenData, err := helpers.CreateAndReturnSiteJoinToken(env, siteName, time.Time{})
			if err != nil {
				helpers.LogTestOutcome(t, condition, err)
			}

			err = sendJoinRequest(env, tokenData)
			if err != nil {
				helpers.LogTestOutcome(t, condition, err)
			}

			err = approveJoinRequest(env, siteName)
			if err != nil {
				helpers.LogTestOutcome(t, condition, err)
			}

			err = sendStatusUpdateInvalidCert(env, tokenData)
			if err != nil && err.Error() == "site not found" {
				err = nil
			}

			helpers.LogTestOutcome(t, condition, err)
		}
	}
}

// sendStatusUpdateNoCert sends a status update to the site manager with no client certificate.
func sendStatusUpdateNoCert(env *helpers.Environment, tokenData types.ExternalSiteTokenBody) error {
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

	path := api.NewURL().Scheme("https").Host(tokenData.Addresses[0]).Path("1.0", "sites", "status")
	return tlsClient.Query(ctx, http.MethodPost, path, nil, nil, nil)
}

// sendStatusUpdateInvalidCert sends a status update to the site manager with a client certificate that was not sent with the join request.
func sendStatusUpdateInvalidCert(env *helpers.Environment, tokenData types.ExternalSiteTokenBody) error {
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

	// send cluster cert as client cert, this should cause site manager to not find the site
	tlsClient, err := helpers.NewTLSHTTPClient(api.URL{}, clusterCert, clusterCertPublicKey)
	if err != nil {
		return err
	}

	path := api.NewURL().Scheme("https").Host(tokenData.Addresses[0]).Path("1.0", "sites", "status")
	return tlsClient.Query(ctx, http.MethodPost, path, nil, nil, nil)
}

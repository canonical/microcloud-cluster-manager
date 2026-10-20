package main

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/canonical/microcloud-cluster-manager/internal/app/cluster-connector/core/auth"
	"github.com/canonical/microcloud-cluster-manager/internal/app/cluster-connector/core/certificate"
	"github.com/canonical/microcloud-cluster-manager/test/helpers"
)

func testMtlsAuthenticator_CacheStartsExpired() (testName string, testFunc func(t *testing.T)) {
	return "mTLS authenticator builds the certificate cache on the first request", func(t *testing.T) {
		var condition string

		{
			condition = "Should start with an expired certificate cache so it is loaded from the database"

			var err error
			authenticator := auth.NewMtlsAuthenticator(nil)
			if !authenticator.Cache().Expired() {
				err = errors.New("new certificate cache is not expired, so enrolled clusters are rejected until the TTL runs out")
			}

			helpers.LogTestOutcome(t, condition, err)
		}
	}
}

func testCertificatesCache_RebuildSkippedWhenFresh() (testName string, testFunc func(t *testing.T)) {
	return "certificate cache rebuild is skipped while the cache is fresh", func(t *testing.T) {
		var condition string

		{
			condition = "Should not query the database when another request already rebuilt the cache"

			cache := &certificate.CertificatesCache{
				Certificates: make(map[string]*certificate.CertificateCacheEntry),
				TTL:          time.Now().Add(time.Minute),
			}

			// The database is nil, so the rebuild panics if it tries to query it.
			err := func() (err error) {
				defer func() {
					if r := recover(); r != nil {
						err = fmt.Errorf("rebuild queried the database: %v", r)
					}
				}()

				return cache.RebuildCache(context.Background(), nil)
			}()

			helpers.LogTestOutcome(t, condition, err)
		}
	}
}

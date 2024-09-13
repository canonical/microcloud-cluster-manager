package state

import (
	"context"
	"crypto/x509"
	"database/sql"
	"sync"
	"time"

	"github.com/canonical/lxd/shared"
	"github.com/canonical/lxd/shared/logger"
	"github.com/canonical/microcluster/v2/rest/types"
	"github.com/canonical/microcluster/v2/state"

	"github.com/canonical/lxd-cluster-manager/internal/database"
)

// CertificateCacheEntry represents a cache entry mapped to a certificate fingerprint.
type CertificateCacheEntry struct {
	ClusterID   int64
	Certificate *x509.Certificate
}

// CertificatesCache represent a cache of LXD cluster certificates with a TTL.
type CertificatesCache struct {
	Certificates map[string]*CertificateCacheEntry
	// TTL is the time when the cache will expire.
	// The cache TTL is used to eliminate the need to synchronize the cache across all members of the cluster.
	// The cache will be re-built using db data after TTL is reached
	TTL time.Time
	mu  sync.RWMutex
}

// AddCertificate adds a certificate entry to the cache.
func (c *CertificatesCache) AddCertificate(cert *x509.Certificate, clusterID int64) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	fingerprint := shared.CertFingerprint(cert)

	c.Certificates[fingerprint] = &CertificateCacheEntry{
		ClusterID:   clusterID,
		Certificate: cert,
	}

	return nil
}

// Expired returns true if the cache has expired.
// No need to lock the cache here since TTL will not change concurrently after cache is created.
func (c *CertificatesCache) Expired() bool {
	return time.Now().After(c.TTL)
}

// GetCertificateEntry returns the cluster ID associated with a certificate fingerprint.
func (c *CertificatesCache) GetCertificateEntry(fingerprint string) (CertificateCacheEntry, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, ok := c.Certificates[fingerprint]
	if !ok || entry == nil {
		return CertificateCacheEntry{}, false
	}

	return *entry, true
}

// GetTrustedCerts returns a map of trusted certificates keyed by their fingerprint.
func (c *CertificatesCache) GetTrustedCerts() map[string]x509.Certificate {
	c.mu.RLock()
	defer c.mu.RUnlock()

	trustedCerts := make(map[string]x509.Certificate)
	for fingerprint, entry := range c.Certificates {
		trustedCerts[fingerprint] = *entry.Certificate
	}

	return trustedCerts
}

// RebuildCache rebuilds the cache from the database.
func (c *CertificatesCache) RebuildCache(ctx context.Context, clutserState state.State) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	return clutserState.Database().Transaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
		var dbRemoteClusters []database.CoreRemoteClusterWithDetail
		var err error

		dbRemoteClusters, err = database.GetCoreRemoteClustersWithDetails(ctx, tx)
		if err != nil {
			return err
		}

		newCacheEntries := make(map[string]*CertificateCacheEntry)
		for _, dbRemoteCluster := range dbRemoteClusters {
			clusterCert, err := types.ParseX509Certificate(dbRemoteCluster.ClusterCertificate)
			if err != nil {
				logger.Warn("Failed to parse remote cluster certificate", logger.Ctx{"remoteCluster": dbRemoteCluster.Name, "err": err})
				continue
			}

			cacheEntry := &CertificateCacheEntry{
				ClusterID:   dbRemoteCluster.ID,
				Certificate: clusterCert.Certificate,
			}

			clusterCertFingerprint := shared.CertFingerprint(clusterCert.Certificate)
			newCacheEntries[clusterCertFingerprint] = cacheEntry
		}

		c.Certificates = newCacheEntries
		c.TTL = time.Now().Add(60 * time.Second)

		return nil
	})
}

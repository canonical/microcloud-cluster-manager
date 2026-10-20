package auth

import (
	"context"
	"fmt"
	"net/http"

	"github.com/canonical/lxd/lxd/request"
	"github.com/canonical/lxd/lxd/util"
	"github.com/canonical/microcloud-cluster-manager/internal/app/cluster-connector/core/certificate"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/database"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/logger"
)

// CtxRemoteClusterID is the context key for the remote cluster ID.
const CtxRemoteClusterID request.CtxKey = "remote-cluster-id"

// MtlsAuthenticator is a mutual TLS authenticator.
type MtlsAuthenticator struct {
	cache *certificate.CertificatesCache
	db    *database.DB
}

// NewMtlsAuthenticator returns a new MtlsAuthenticator.
func NewMtlsAuthenticator(db *database.DB) *MtlsAuthenticator {
	return &MtlsAuthenticator{
		// The cache TTL is left unset, so the cache starts out expired and is built from the database on the first
		// request. Otherwise clusters that are already enrolled are rejected until the first TTL runs out.
		cache: &certificate.CertificatesCache{
			Certificates: make(map[string]*certificate.CertificateCacheEntry),
		},
		db: db,
	}
}

// Auth authenticates a request using mutual TLS.
func (ma *MtlsAuthenticator) Auth(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	if r.TLS == nil || len(r.TLS.PeerCertificates) == 0 {
		return fmt.Errorf("tls is required")
	}

	if len(r.TLS.PeerCertificates) != 1 {
		return fmt.Errorf("expected exactly one peer certificate")
	}

	if ma.cache.Expired() {
		err := ma.cache.RebuildCache(ctx, ma.db)

		if err != nil {
			return fmt.Errorf("failed to rebuild cache: %w", err)
		}
	}

	peerCert := r.TLS.PeerCertificates[0]
	trustedCerts := ma.cache.GetTrustedCerts()
	trusted, fingerprint := util.CheckMutualTLS(*peerCert, trustedCerts)

	if !trusted {
		logger.Log.Info("AUTHN untrusted peer certificate presented for mTLS")
		return fmt.Errorf("invalid cluster certificate")
	}

	// The cache may have been rebuilt since GetTrustedCerts was called, and no longer contain this certificate.
	remoteClusterCert, ok := ma.cache.GetCertificateEntry(fingerprint)
	if !ok {
		logger.Log.Info("AUTHN peer certificate for mTLS was removed from the trusted certificates")
		return fmt.Errorf("invalid cluster certificate")
	}

	request.SetContextValue(r, CtxRemoteClusterID, remoteClusterCert.ClusterID)

	logger.Log.Info("AUTHN peer certificate for mTLS authenticated successfully")
	return nil
}

// Cache returns the certificates cache.
func (ma *MtlsAuthenticator) Cache() *certificate.CertificatesCache {
	return ma.cache
}

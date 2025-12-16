package auth

import (
	"context"
	"fmt"
	"net/http"
	"time"

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
		cache: &certificate.CertificatesCache{
			Certificates: make(map[string]*certificate.CertificateCacheEntry),
			TTL:          time.Now().Add(60 * time.Second),
		},
		db: db,
	}
}

// Auth authenticates a request using mutual TLS.
func (ma *MtlsAuthenticator) Auth(ctx context.Context, w http.ResponseWriter, r *http.Request) (bool, error) {
	if r.TLS == nil || len(r.TLS.PeerCertificates) == 0 {
		return false, fmt.Errorf("tls is required")
	}

	if len(r.TLS.PeerCertificates) != 1 {
		return false, fmt.Errorf("expected exactly one peer certificate")
	}

	if ma.cache.Expired() {
		err := ma.cache.RebuildCache(ctx, ma.db)

		if err != nil {
			return false, err
		}
	}

	peerCert := r.TLS.PeerCertificates[0]
	trustedCerts := ma.cache.GetTrustedCerts()
	trusted, fingerprint := util.CheckMutualTLS(*peerCert, trustedCerts)

	if !trusted {
		logger.Log.Info("AUTHN untrusted peer certificate presented for mTLS")
		return false, fmt.Errorf("invalid cluster certificate")
	}

	remoteClusterCert, _ := ma.cache.GetCertificateEntry(fingerprint)
	request.SetContextValue(r, CtxRemoteClusterID, remoteClusterCert.ClusterID)

	logger.Log.Info("AUTHN peer certificate for mTLS authenticated successfully")
	return true, nil
}

// Cache returns the certificates cache.
func (ma *MtlsAuthenticator) Cache() *certificate.CertificatesCache {
	return ma.cache
}

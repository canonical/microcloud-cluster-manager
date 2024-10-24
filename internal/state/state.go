package state

import (
	"sync"
	"time"

	"github.com/canonical/microcluster/v2/microcluster"

	"github.com/canonical/lxd-cluster-manager/internal/database"
	"github.com/canonical/lxd-cluster-manager/internal/oidc"
)

// ClusterManagerState holds the global state of the Cluster Manager.
type ClusterManagerState struct {
	// OIDCVerifier is the OpenID Connect verifier used for user authentication and validate authentication for protected API endpoints.
	OIDCVerifier *oidc.Verifier
	// CertificateCache is a map of certificate fingerprints to remote cluster certificates.
	CertificatesCache *CertificatesCache
	mu                sync.RWMutex
	Database          database.Database
}

// New creates a new ClusterManagerState.
func New(m *microcluster.MicroCluster) *ClusterManagerState {
	return &ClusterManagerState{
		CertificatesCache: &CertificatesCache{
			Certificates: make(map[string]*CertificateCacheEntry),
			TTL:          time.Now().Add(60 * time.Second),
		},
	}
}

// SetOIDCVerifier sets the OIDCVerifier.
func (s *ClusterManagerState) SetOIDCVerifier(verifier *oidc.Verifier) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.OIDCVerifier = verifier
}

// SetDatabase sets the database.
func (s *ClusterManagerState) SetDatabase(db database.Database) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.Database = db
}

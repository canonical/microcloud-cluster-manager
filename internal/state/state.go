package state

import (
	"sync"

	"github.com/canonical/microcluster/microcluster"

	"github.com/canonical/lxd-cluster-manager/internal/oidc"
)

// ClusterManagerState holds the global state of the Cluster Manager.
type ClusterManagerState struct {
	// OIDCVerifier is the OpenID Connect verifier used for user authentication and validate authentication for protected API endpoints.
	OIDCVerifier *oidc.Verifier
	mu           sync.RWMutex
}

// New creates a new ClusterManagerState.
func New(m *microcluster.MicroCluster) *ClusterManagerState {
	return &ClusterManagerState{}
}

// SetOIDCVerifier sets the OIDCVerifier.
func (s *ClusterManagerState) SetOIDCVerifier(verifier *oidc.Verifier) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.OIDCVerifier = verifier
}

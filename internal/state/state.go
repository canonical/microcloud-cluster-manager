package state

import (
	"github.com/canonical/microcluster/microcluster"
)

// SiteManagerState holds the global state of the site manager.
type SiteManagerState struct {
	// MicroCluster is the MicroCluster instance.
	MicroCluster *microcluster.MicroCluster
}

// New creates a new SiteManagerState.
func New(m *microcluster.MicroCluster) *SiteManagerState {
	return &SiteManagerState{
		MicroCluster: m,
	}
}

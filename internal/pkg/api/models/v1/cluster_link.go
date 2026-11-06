package models

import "github.com/canonical/lxd/shared/api"

// ClusterLinkGroup holds the cluster links defined on one remote cluster.
type ClusterLinkGroup struct {
	Cluster string            `json:"cluster"`
	Links   []api.ClusterLink `json:"links"`
	Error   string            `json:"error,omitempty"`
}

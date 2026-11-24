package models

import (
	"github.com/canonical/lxd/shared/api"
)

// ResolvedClusterLink is a cluster link with the name of its source cluster, the managed cluster it is defined on,
// and the name of its target cluster, the managed cluster it points at. The target cluster is nil when the link
// does not point at a managed cluster.
type ResolvedClusterLink struct {
	Link          api.ClusterLink `json:"link"`
	SourceCluster string          `json:"source_cluster"`
	TargetCluster *string         `json:"target_cluster"`
}

// ClusterLinkError reports a cluster whose cluster links could not be loaded.
type ClusterLinkError struct {
	Cluster string `json:"cluster"`
	Error   string `json:"error"`
}

// ClusterLinkList is the response of the cluster link list endpoint, for one cluster or for all clusters. Clusters
// whose links could not be loaded are reported in the errors next to the links that could be loaded.
type ClusterLinkList struct {
	ClusterLinks []ResolvedClusterLink `json:"cluster_links"`
	Errors       []ClusterLinkError    `json:"errors"`
}

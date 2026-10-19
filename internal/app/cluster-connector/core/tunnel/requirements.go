package tunnel

import "github.com/canonical/lxd/shared/api"

// ClusterRequirements describes the capabilities and permissions required for an operation on a cluster.
// Empty permissions still require reachability and the listed API extensions.
type ClusterRequirements struct {
	APIExtensions []string
	Permissions   []api.Permission
}

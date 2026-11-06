// Package replicator holds the replicator rules shared by the API handlers: the LXD API extension
// a cluster must support, the supported run actions and the project-scoped LXD API paths.
// Permissions are enforced by LXD itself on each request.
package replicator

import (
	"fmt"
	"net/url"

	"github.com/canonical/microcloud-cluster-manager/internal/app/cluster-connector/core/tunnel"
)

// APIExtension is the LXD API extension that adds replicator support.
const APIExtension = "replicators"

// Supported actions to run a replicator.
const (
	ActionStart   = "start"
	ActionRestore = "restore"
)

// Requirements returns what a cluster must support to manage its replicators.
func Requirements() tunnel.ClusterRequirements {
	return tunnel.ClusterRequirements{
		APIExtensions: []string{APIExtension},
	}
}

// ValidateAction returns an error if the given replicator run action is not supported.
func ValidateAction(action string) error {
	switch action {
	case ActionStart, ActionRestore:
		return nil
	default:
		return fmt.Errorf("Unsupported replicator action %q", action)
	}
}

// CollectionPath returns the LXD API path of the replicators in the given project.
func CollectionPath(project string) string {
	return "/1.0/replicators?" + projectQuery(project).Encode()
}

// ListPath returns the LXD API path listing the full replicators in the given project,
// or in all projects if no project is given.
func ListPath(project string) string {
	query := url.Values{}
	if project == "" {
		query.Set("all-projects", "true")
	} else {
		query.Set("project", project)
	}

	query.Set("recursion", "2")
	return "/1.0/replicators?" + query.Encode()
}

// Path returns the LXD API path of a replicator in the given project.
func Path(project string, name string) string {
	return "/1.0/replicators/" + url.PathEscape(name) + "?" + projectQuery(project).Encode()
}

// StatePath returns the LXD API path of the state of a replicator in the given project.
func StatePath(project string, name string) string {
	return "/1.0/replicators/" + url.PathEscape(name) + "/state?" + projectQuery(project).Encode()
}

func projectQuery(project string) url.Values {
	query := url.Values{}
	query.Set("project", project)
	return query
}

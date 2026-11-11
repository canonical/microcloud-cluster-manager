// Package project holds the LXD API paths used by the project API handlers.
// Permissions are enforced by LXD itself on each request.
package project

import (
	"net/url"
)

// ListPath returns the LXD API path listing the full projects of a cluster.
func ListPath() string {
	query := url.Values{}
	query.Set("recursion", "1")
	return "/1.0/projects?" + query.Encode()
}

// CreatePath returns the LXD API path to create a project.
func CreatePath() string {
	return "/1.0/projects"
}

// Path returns the LXD API path of the given project.
func Path(name string) string {
	return "/1.0/projects/" + url.PathEscape(name)
}

// StatePath returns the LXD API path of the state of the given project.
// When force is set, LXD skips the replication topology guards on a state update.
func StatePath(name string, force bool) string {
	path := Path(name) + "/state"
	if force {
		query := url.Values{}
		query.Set("force", "true")
		path += "?" + query.Encode()
	}

	return path
}

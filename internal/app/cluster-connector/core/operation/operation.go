// Package operation holds the LXD API paths used by the operation API handlers.
// Permissions are enforced by LXD itself on each request.
package operation

import (
	"net/url"
)

// Path returns the LXD API path of the given operation.
func Path(id string) string {
	return "/1.0/operations/" + url.PathEscape(id)
}

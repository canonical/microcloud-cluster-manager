package types

import "github.com/canonical/microcluster/rest/types"

const (
	// APIVersionPrefix is the path prefix for API related endpoints.
	APIVersionPrefix types.EndpointPrefix = "1.0"
	// NoPrefix is the path prefix for any endpoints that should be located at server root.
	NoPrefix types.EndpointPrefix = ""
	// InternalPublicEndpoint is restricted to trusted servers.
	InternalPublicEndpoint types.EndpointPrefix = "core/1.0"
)

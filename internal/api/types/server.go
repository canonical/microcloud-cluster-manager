package types

import "github.com/canonical/microcluster/rest/types"

const (
	// APIVersionPrefix is the path prefix for API related endpoints.
	APIVersionPrefix types.EndpointPrefix = "1.0"
	// NoPrefix is the path prefix for any endpoints that should be located at server root.
	NoPrefix types.EndpointPrefix = ""
	// InternalEndpoint is restricted to trusted servers.
	InternalEndpoint types.EndpointPrefix = "cluster/internal"
)

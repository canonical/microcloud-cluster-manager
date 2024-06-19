package types

import "github.com/canonical/microcluster/rest/types"

const (
	// ApiVersionPrefix is the path prefix for API related endpoints.
	ApiVersionPrefix types.EndpointPrefix = "1.0"
	// NoPrefix is the path prefix for any endpoints that should be located at server root.
	NoPrefix types.EndpointPrefix = ""
)

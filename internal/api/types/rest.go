package types

import (
	"net/http"

	"github.com/canonical/lxd/lxd/response"
	"github.com/canonical/microcluster/state"
)

// AccessHandler represents a function that handles an API endpoint with access control.
type AccessHandler func(state state.State, r *http.Request) (bool, response.Response)

// EndpointHandler represents a function that handles an API endpoint.
type EndpointHandler func(state.State, *http.Request) response.Response

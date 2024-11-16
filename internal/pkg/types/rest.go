package types

import (
	"net/http"

	"github.com/canonical/lxd-cluster-manager/internal/pkg/database"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// EndpointHandler is a function that returns a http.HandlerFunc
type EndpointHandler func(w http.ResponseWriter, r *http.Request) error

// Endpoint holds the handler function, method and path for a route
type Endpoint struct {
	Handler func(*database.DB, *zap.SugaredLogger) EndpointHandler
	Method  string
	Path    string
}

// RouteGroup holds a prefix and a list of endpoints
type RouteGroup struct {
	IsRoot      bool
	Prefix      string
	Endpoints   []Endpoint
	Middlewares []mux.MiddlewareFunc
}

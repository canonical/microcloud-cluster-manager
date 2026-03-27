package types

import (
	"context"
	"net/http"

	"github.com/canonical/microcloud-cluster-manager/internal/pkg/config"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/database"
	"github.com/gorilla/mux"
)

// EndpointHandler is a function that returns a http.HandlerFunc.
type EndpointHandler func(w http.ResponseWriter, r *http.Request) error

// Endpoint holds the handler function, method and path for a route.
type Endpoint struct {
	Handler             func(RouteConfig) EndpointHandler
	Method              string
	Path                string
	AllowUnauthorized   bool
	AllowedEntitlements []string
}

// Authenticator represents the interface that each service in cluster manager must implement for securing their respective APIs.
type Authenticator interface {
	Auth(ctx context.Context, w http.ResponseWriter, r *http.Request) (bool, error)
}

// Authorizor represent the interface that each service in cluster manager must implement for enforcing authorization based on entitlements.
type Authorizor interface {
	CheckPermissions(ctx context.Context, allowedEntitlements []string) error
}

// RateLimiter represent the interface that each service in cluster manager must implement for enforcing rate limit.
type RateLimiter interface {
	CheckLimit(ctx context.Context, w http.ResponseWriter, r *http.Request) (bool, error)
}

// RouteConfig holds the necessary dependencies for routes and middlewares within service APIs.
type RouteConfig struct {
	Auth        Authenticator
	Authorizor  Authorizor
	RateLimiter RateLimiter
	DB          *database.DB
	Env         *config.Config
}

// RouteMiddleware represents middlewares in service APIs that requires route dependencies.
type RouteMiddleware func(RouteConfig) mux.MiddlewareFunc

// RouteGroup holds a prefix and a list of endpoints.
type RouteGroup struct {
	IsRoot      bool
	Prefix      string
	Endpoints   []Endpoint
	Middlewares []RouteMiddleware
}

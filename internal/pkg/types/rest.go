package types

import (
	"context"
	"net/http"
	"sync"

	"github.com/canonical/microcloud-cluster-manager/internal/pkg/config"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/database"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
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

type Auth struct {
	Authenticator Authenticator
	Authorizor    Authorizor
}

// Authenticator represents the interface that each service in cluster manager must implement for securing their respective APIs.
type Authenticator interface {
	Auth(ctx context.Context, w http.ResponseWriter, r *http.Request) error
}

// Authorizor represents the interface that each service in cluster manager must implement for enforcing authorization based on entitlements.
type Authorizor interface {
	CheckPermissions(ctx context.Context, allowedEntitlements []string) error
}

// RateLimiter represents the interface that each service in cluster manager must implement for enforcing rate limit.
type RateLimiter interface {
	CheckLimit(ctx context.Context, w http.ResponseWriter, r *http.Request) (bool, error)
}

// RouteConfig holds the necessary dependencies for routes and middlewares within service APIs.
type RouteConfig struct {
	Auth        Auth
	RateLimiter RateLimiter
	DB          *database.DB
	Env         *config.Config
	TunnelStore *TunnelStore
}

// TunnelStore manages WebSocket connections for different MicroClouds.
type TunnelStore struct {
	Mu              sync.RWMutex
	TunnelByCluster map[int]*Tunnel
}

// Tunnel is the reverse tunnel with a specific MicroCloud.
type Tunnel struct {
	Mu               sync.RWMutex
	WsConn           *websocket.Conn
	PendingResponses map[string]chan ClusterManagerTunnelResponse
	UserSessions     map[string]string
}

// ClusterManagerTunnelRequest is forwarded over reverse tunnel.
type ClusterManagerTunnelRequest struct {
	UUID    string      `json:"uuid"`
	Method  string      `json:"method"`
	Path    string      `json:"path"`
	Headers http.Header `json:"headers"`
	Body    []byte      `json:"body"`
}

// ClusterManagerTunnelResponse is received from to the reverse tunnel.
type ClusterManagerTunnelResponse struct {
	UUID    string         `json:"uuid"`
	Status  int            `json:"status"`
	Headers http.Header    `json:"headers"`
	Cookies []*http.Cookie `json:"cookies"`
	Body    []byte         `json:"body"`
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

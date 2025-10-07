package api

import (
	"context"
	"net/http"
	"os"
	"strings"
	"syscall"

	"github.com/canonical/lxd/lxd/response"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/config"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/database"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/logger"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/types"
	"github.com/gorilla/mux"
)

// A Handler is a type that handles an http request within our own little mini
// framework.
type Handler func(ctx context.Context, w http.ResponseWriter, r *http.Request) error

// APIConfig is used to configure the API which represents the entry point of any service in the application.
type APIConfig struct {
	Shutdown  chan os.Signal
	DB        *database.DB
	Auth      types.Authenticator
	EnvConfig *config.Config
}

// API is the entrypoint into our application and what configures our context
// object for each of our http handlers.
type API struct {
	mux       *mux.Router
	shutdown  chan os.Signal
	db        *database.DB
	auth      types.Authenticator
	envConfig *config.Config
}

// NewAPI creates a new API.
func NewAPI(cfg APIConfig) *API {
	mux := mux.NewRouter()
	mux.StrictSlash(false)
	mux.SkipClean(true)
	mux.UseEncodedPath()

	return &API{
		mux:       mux,
		shutdown:  cfg.Shutdown,
		db:        cfg.DB,
		auth:      cfg.Auth,
		envConfig: cfg.EnvConfig,
	}
}

// SignalShutdown is used to gracefully shutdown the app when an integrity
// issue is identified.
func (a *API) SignalShutdown() {
	a.shutdown <- syscall.SIGTERM
}

// UseGlobalMiddleWares adds global middlewares to the router.
func (a *API) UseGlobalMiddleWares(mw ...mux.MiddlewareFunc) {
	a.mux.Use(mw...)
}

// RegisterRoutes adds the routes to the router.
func (a *API) RegisterRoutes(routes []types.RouteGroup) {
	rc := types.RouteConfig{
		Auth: a.auth,
		DB:   a.db,
		Env:  a.envConfig,
	}

	registerRoutes(a.mux, routes, rc)

	a.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		err := response.SyncResponse(true, []string{"/1.0"}).Render(w, r)
		if err != nil {
			logger.Log.Errorw("Failed to write HTTP response", "url", r.URL, "err", err.Error())
		}
	})

	a.mux.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Log.Infow("Sending top level 404", "url", r.URL)
		w.Header().Set("Content-Type", "application/json")
		err := response.NotFound(nil).Render(w, r)
		if err != nil {
			logger.Log.Error("Failed to write HTTP response", "url", r.URL, "err", err.Error())
		}
	})
}

// Mux returns the router.
func (a *API) Mux() *mux.Router {
	return a.mux
}

// ServeHTTP implements the http.Handler interface.
func (a *API) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.mux.ServeHTTP(w, r)
}

// GetStatusServer sets up the status endpoint for liveliness and readiness checks.
func (a *API) GetStatusServer(port string) *http.Server {
	// Define the status server
	statusRouter := mux.NewRouter()
	statusRouter.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		err := a.db.StatusCheck(r.Context())
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			logger.Log.Errorw("Status check failed", "err", err)
			return
		}

		w.WriteHeader(http.StatusOK)
		logger.Log.Infow("Status ok")
	})

	statusServer := &http.Server{
		Addr:    ":" + port, // Status server listens on a different port
		Handler: statusRouter,
	}

	return statusServer
}

// isBrowserClient checks if the request is coming from a browser client.
func isBrowserClient(r *http.Request) bool {
	// Check if the User-Agent starts with "Mozilla" which is common for browsers.
	return strings.HasPrefix(r.Header.Get("User-Agent"), "Mozilla")
}

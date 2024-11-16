package api

import (
	"context"
	"net/http"
	"os"
	"syscall"

	"github.com/gorilla/mux"
	"go.uber.org/zap"

	"github.com/canonical/lxd-cluster-manager/internal/pkg/database"
	"github.com/canonical/lxd-cluster-manager/internal/pkg/types"
	"github.com/canonical/lxd/lxd/response"
)

// Router represents a group of related routes
type Router interface {
	RegisterRoutes(r *mux.Router, db *database.DB)
}

// A Handler is a type that handles an http request within our own little mini
// framework.
type Handler func(ctx context.Context, w http.ResponseWriter, r *http.Request) error

type ApiConfig struct {
	Shutdown chan os.Signal
	DB       *database.DB
	Logger   *zap.SugaredLogger
}

// Api is the entrypoint into our application and what configures our context
// object for each of our http handlers.
type Api struct {
	mux      *mux.Router
	shutdown chan os.Signal
	db       *database.DB
	logger   *zap.SugaredLogger
}

// NewApp creates an App value that handle a set of routes for the application.
func NewApi(cfg ApiConfig) *Api {
	mux := mux.NewRouter()
	mux.StrictSlash(false)
	mux.SkipClean(true)
	mux.UseEncodedPath()

	return &Api{
		mux:      mux,
		shutdown: cfg.Shutdown,
		db:       cfg.DB,
		logger:   cfg.Logger,
	}
}

// SignalShutdown is used to gracefully shutdown the app when an integrity
// issue is identified.
func (a *Api) SignalShutdown() {
	a.shutdown <- syscall.SIGTERM
}

// UseGlobalMiddleWares adds global middlewares to the router.
func (a *Api) UseGlobalMiddleWares(mw ...mux.MiddlewareFunc) {
	a.mux.Use(mw...)
}

// RegisterRoutes adds the routes to the router.
func (a *Api) RegisterRoutes(routes []types.RouteGroup, version string) {
	registerRoutes(a.mux, a.db, a.logger, routes, version)

	a.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		err := response.SyncResponse(true, []string{"/1.0"}).Render(w, r)
		if err != nil {
			a.logger.Errorw("Failed to write HTTP response", "url", r.URL, "err", err.Error())
		}
	})

	a.mux.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		a.logger.Infow("Sending top level 404", "url", r.URL)
		w.Header().Set("Content-Type", "application/json")
		err := response.NotFound(nil).Render(w, r)
		if err != nil {
			a.logger.Error("Failed to write HTTP response", "url", r.URL, "err", err.Error())
		}
	})
}

// ServeHTTP implements the http.Handler interface.
func (a *Api) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.mux.ServeHTTP(w, r)
}

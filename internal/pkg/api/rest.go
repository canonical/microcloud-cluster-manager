package api

import (
	"net/http"
	"path"

	"github.com/canonical/lxd-cluster-manager/internal/pkg/database"
	"github.com/canonical/lxd-cluster-manager/internal/pkg/types"
	"github.com/canonical/lxd/lxd/response"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

func registerRoutes(mux *mux.Router, db *database.DB, logger *zap.SugaredLogger, routes []types.RouteGroup, version string) {
	// Register route groups
	for _, r := range routes {
		registerRouteGroup(mux, db, logger, r, version)
	}
}

func registerRouteGroup(mux *mux.Router, db *database.DB, logger *zap.SugaredLogger, r types.RouteGroup, version string) {
	routeGroupPath := path.Join("/", r.Prefix)
	if !r.IsRoot {
		routeGroupPath = path.Join("/", version, r.Prefix)
	}

	// apply middlewares at route group level
	sr := mux.PathPrefix(routeGroupPath).Subrouter()
	if len(r.Middlewares) > 0 {
		sr.Use(r.Middlewares...)
	}

	for _, e := range r.Endpoints {
		registerEndpoint(sr, db, logger, routeGroupPath, e)
	}
}

func registerEndpoint(mux *mux.Router, db *database.DB, logger *zap.SugaredLogger, prefix string, e types.Endpoint) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		err := e.Handler(db, logger)(w, r)
		if err != nil {
			logger.Errorw("internal error", "ERROR", err)
			renderErr := response.InternalError(err).Render(w, r)
			if renderErr != nil {
				logger.Errorw("failed to write error response", "path", path.Join(prefix, e.Path), "ERROR", renderErr.Error())
			}
		}
	})

	// in case if the endpoint is the root of the route group
	ep := ""
	if e.Path != "" {
		ep = path.Join("/", e.Path)
	}

	mux.Handle(ep, handler).Methods(e.Method)
}

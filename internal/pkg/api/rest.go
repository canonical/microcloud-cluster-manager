package api

import (
	"net/http"
	"path"
	"slices"

	"github.com/canonical/lxd/lxd/response"
	"github.com/canonical/lxd/shared"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/logger"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/types"
	"github.com/gorilla/mux"
)

// secFetchSiteForbidden defines client Sec-Fetch-Site header values that will be forbidden access.
var secFetchSiteForbidden = []string{"cross-site", "same-site"}

func registerRoutes(mux *mux.Router, routes []types.RouteGroup, rc types.RouteConfig) {
	// Register route groups
	for _, rg := range routes {
		registerRouteGroup(mux, rg, rc)
	}
}

func registerRouteGroup(mux *mux.Router, rg types.RouteGroup, rc types.RouteConfig) {
	routeGroupPath := path.Join("/", rg.Prefix)
	if !rg.IsRoot {
		routeGroupPath = path.Join("/", rc.Env.APIVersion, rg.Prefix)
	}

	// apply middlewares at route group level
	sr := mux.PathPrefix(routeGroupPath).Subrouter()
	if len(rg.Middlewares) > 0 {
		for _, m := range rg.Middlewares {
			sr.Use(m(rc))
		}
	}

	for _, e := range rg.Endpoints {
		registerEndpoint(sr, routeGroupPath, e, rc)
	}
}

func registerEndpoint(mux *mux.Router, prefix string, e types.Endpoint, rc types.RouteConfig) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// Protect against CSRF when using UI with browser that supports Fetch metadata.
		// Deny Sec-Fetch-Site when set to cross-site or same-site.
		allowedPaths := []string{"/oidc/callback", "/ui", "/ui/"}
		if slices.Contains(secFetchSiteForbidden, r.Header.Get("Sec-Fetch-Site")) && !slices.Contains(allowedPaths, r.URL.Path) {
			renderErr := response.ErrorResponse(http.StatusForbidden, "Forbidden Sec-Fetch-Site header value").Render(w, r)
			if renderErr != nil {
				logger.Log.Errorw("failed to write error response", "path", path.Join(prefix, e.Path), "ERROR", renderErr.Error())
			}
			return
		}

		// Validate browser Content-Type if supplied, or if non-zero Content-Length supplied.
		if isBrowserClient(r) {
			contentTypeParts := shared.SplitNTrimSpace(r.Header.Get("Content-Type"), ";", 2, false) // Ignore multi-part boundary part.
			contentLength := r.Header.Get("Content-Length")
			hasContentLength := contentLength != "" && contentLength != "0"
			if (hasContentLength || contentTypeParts[0] != "") && !slices.Contains([]string{"application/json"}, contentTypeParts[0]) {
				renderErr := response.ErrorResponse(http.StatusUnsupportedMediaType, "Unsupported Content-Type for this request").Render(w, r)
				if renderErr != nil {
					logger.Log.Errorw("failed to write error response", "path", path.Join(prefix, e.Path), "ERROR", renderErr.Error())
				}
				return
			}
		}

		if !e.AllowUnauthorized {
			err := rc.Authorizor.CheckPermissions(r.Context(), e.AllowedEntitlements)
			if err != nil {
				renderErr := response.ErrorResponse(http.StatusForbidden, err.Error()).Render(w, r)
				if renderErr != nil {
					logger.Log.Errorw("failed to write error response", "path", path.Join(prefix, e.Path), "ERROR", renderErr.Error())
				}
				return
			}
		}

		err := e.Handler(rc)(w, r)
		if err != nil {
			logger.Log.Errorw("internal error", "ERROR", err)
			renderErr := response.InternalError(err).Render(w, r)
			if renderErr != nil {
				logger.Log.Errorw("failed to write error response", "path", path.Join(prefix, e.Path), "ERROR", renderErr.Error())
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

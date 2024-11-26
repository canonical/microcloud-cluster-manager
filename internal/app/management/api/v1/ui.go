package v1

import (
	"embed"
	"io/fs"
	"net/http"
	"os"

	"github.com/canonical/lxd-cluster-manager/internal/pkg/types"
	"github.com/canonical/lxd/lxd/response"
)

// UI files are copied to the static directory. Then embed the static directory in go binary.
//
//go:generate cp -r ../../../../../ui/build/ui ./static
//go:embed static
var UI_FS embed.FS

var UI = types.RouteGroup{
	IsRoot: true,
	Prefix: "",
	Endpoints: []types.Endpoint{
		{
			Path:    "ui/{path:.*}",
			Method:  http.MethodGet,
			Handler: serveUI,
		},
	},
}

var UIRoot = types.RouteGroup{
	IsRoot: true,
	Prefix: "",
	Endpoints: []types.Endpoint{
		{
			Path:    "ui",
			Method:  http.MethodGet,
			Handler: redirectToUI,
		},
		{
			Path:    "",
			Method:  http.MethodGet,
			Handler: redirectToUI,
		},
		{
			Path:    "/",
			Method:  http.MethodGet,
			Handler: redirectToUI,
		},
	},
}

func redirectToUI(rc types.RouteConfig) types.EndpointHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		return response.SyncResponseRedirect("/ui/").Render(w, r)
	}
}

func serveUI(rc types.RouteConfig) types.EndpointHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		uiFS, err := fs.Sub(UI_FS, "static")
		if err != nil {
			return response.InternalError(err).Render(w, r)
		}

		// Need to implement the http.FileSystem interface to serve the UI files.
		// Need custom Open method to serve index.html when the requested file is not found.
		// e.g. /ui/unknown-file -> /ui/index.html
		// This will allow react router to handle the routing to the correct component file.
		uiHTTPDir := uiHTTPDir{http.FS(uiFS)}

		fileServer := http.StripPrefix("/ui", http.FileServer(uiHTTPDir))

		// Content-Type header is set to application/json by default. We need to remove it to serve the UI.
		w.Header().Del("Content-Type")
		// Disables the FLoC (Federated Learning of Cohorts) feature on the browser,
		// preventing the current page from being included in the user's FLoC calculation.
		// FLoC is a proposed replacement for third-party cookies to enable interest-based advertising.
		w.Header().Set("Permissions-Policy", "interest-cohort=()")
		// Prevents the browser from trying to guess the MIME type, which can have security implications.
		// This tells the browser to strictly follow the MIME type provided in the Content-Type header.
		w.Header().Set("X-Content-Type-Options", "nosniff")
		// Restricts the page from being displayed in a frame, iframe, or object to avoid click jacking attacks,
		// but allows it if the site is navigating to the same origin.
		w.Header().Set("X-Frame-Options", "SAMEORIGIN")
		// Sets the Content Security Policy (CSP) for the page, which helps mitigate XSS attacks and data injection attacks.
		// The policy allows loading resources (scripts, styles, images, etc.) only from the same origin ('self'), data URLs, and all subdomains of ubuntu.com.
		w.Header().Set("Content-Security-Policy", "default-src 'self' data: https://*.ubuntu.com https://*.canonical.com; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'")

		fileServer.ServeHTTP(w, r)

		return nil
	}
}

type uiHTTPDir struct {
	http.FileSystem
}

// Open opens the HTTP server for the user interface files.
func (fs uiHTTPDir) Open(name string) (http.File, error) {
	fsFile, err := fs.FileSystem.Open(name)
	if err != nil && os.IsNotExist(err) {
		return fs.FileSystem.Open("index.html")
	}

	return fsFile, err
}

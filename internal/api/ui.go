package api

import (
	"embed"
	"io/fs"
	"net/http"
	"os"

	"github.com/canonical/lxd/lxd/response"
	"github.com/canonical/microcluster/rest"
	"github.com/canonical/microcluster/state"
)

// UI files are copied to the static directory. Then embed the static directory in go binary.
//
//go:generate cp -r ../../ui/build/ui ./static
//go:embed static
var UI embed.FS

var uiRootCmd = rest.Endpoint{
	Path: "",
	Get:  rest.EndpointAction{Handler: redirectToUI, AllowUntrusted: true},
}

var uiServeRoutes = []string{
	"ui",
	"ui/assets/{asset}",
	"ui/assets/img/{image}",
	"ui/login",
	"ui/sites",
	"ui/sites/pending",
	"ui/sites/tokens",
	"ui/settings",
}

func generateUIEndpoints() []rest.Endpoint {
	var uiEndpoints []rest.Endpoint
	for _, route := range uiServeRoutes {
		uiEndpoints = append(uiEndpoints, rest.Endpoint{
			Path: route,
			Get:  rest.EndpointAction{Handler: serveUI, AllowUntrusted: true},
		})
	}

	return uiEndpoints
}

func redirectToUI(s state.State, r *http.Request) response.Response {
	return response.SyncResponseRedirect("/ui")
}

func serveUI(s state.State, r *http.Request) response.Response {
	uiFS, err := fs.Sub(UI, "static")
	if err != nil {
		return response.InternalError(err)
	}

	// Need to implement the http.FileSystem interface to serve the UI files.
	// Need custom Open method to serve index.html when the requested file is not found.
	// e.g. /ui/unknown-file -> /ui/index.html
	// This will allow react router to handle the routing to the correct component file.
	uiHTTPDir := uiHTTPDir{http.FS(uiFS)}

	fileServer := http.StripPrefix("/ui", http.FileServer(uiHTTPDir))

	serverUIHandler := func(w http.ResponseWriter) error {
		// microcluster sets the Content-Type header to application/json by default. We need to remove it to serve the UI.
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

	return response.ManualResponse(serverUIHandler)
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

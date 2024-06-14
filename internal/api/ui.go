package api

import (
	"embed"
	"io/fs"
	"net/http"

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

var uiCmd = rest.Endpoint{
	Path: "ui",
	Get:  rest.EndpointAction{Handler: serveUI, AllowUntrusted: true},
}

var uiAssetsCmd = rest.Endpoint{
	Path: "ui/assets/{asset}",
	Get:  rest.EndpointAction{Handler: serveUI, AllowUntrusted: true},
}

var uiImgCmd = rest.Endpoint{
	Path: "ui/assets/img/{image}",
	Get:  rest.EndpointAction{Handler: serveUI, AllowUntrusted: true},
}

func redirectToUI(s *state.State, r *http.Request) response.Response {
	return response.SyncResponseRedirect("/ui")
}

func serveUI(s *state.State, r *http.Request) response.Response {
	uiFS, err := fs.Sub(UI, "static")
	if err != nil {
		return response.InternalError(err)
	}

	fileServer := http.StripPrefix("/ui", http.FileServer(http.FS(uiFS)))

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

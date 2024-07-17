package types

import "net/http"

// AuthCallbackPost represents the request body for the oidc/callback endpoint for a forwarded request.
type AuthCallbackPost struct {
	Cookies []http.Cookie `json:"cookies"`
}

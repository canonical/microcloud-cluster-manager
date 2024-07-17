package client

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/canonical/lxd/shared/api"
	"github.com/canonical/microcluster/client"

	"github.com/canonical/lxd-site-manager/internal/api/types"
	"github.com/canonical/lxd-site-manager/internal/oidc"
)

// CallbackForwardCmd forwards a oidc callback request to the oidc/callback endpoint on.
func CallbackForwardCmd(ctx context.Context, c *client.Client, sourceRequest *http.Request) (oidc.OidcTokens, error) {
	queryCtx, cancel := context.WithTimeout(ctx, time.Second*30)
	defer cancel()

	// Get all the source query params and append it to the target URL
	sourceQueryParams := sourceRequest.URL.Query()
	targetURL := api.NewURL().Path("oidc", "callback")
	targetQueryParams := targetURL.Query()
	for key, values := range sourceQueryParams {
		for _, value := range values {
			targetQueryParams.Add(key, value)
		}
	}

	// FIXME: microcluster does not allow adjusting requets headers when forwarding requests using built-in client functions
	// This is the reason why we need to set the forwarded query parameter for a forwarded request
	// We should remove this once the microcluster is updated to allow adjusting request headers
	targetQueryParams.Add("forwarded", "true")
	targetURL.RawQuery = targetQueryParams.Encode()

	// Clear the raw path, this causes the endpoint to not be found when forwarding the request
	targetURL.RawPath = ""

	// Get all the source cookies and append it to the payload
	// FIXME: instead of forwarding cookies in the payload, we should forward them in the request headers
	// We should remove this once the microcluster is updated to allow adjusting request headers
	cookies := sourceRequest.Cookies()
	var cookieSlice []http.Cookie
	for _, cookie := range cookies {
		cookieSlice = append(cookieSlice, *cookie)
	}

	payload := types.AuthCallbackPost{
		Cookies: cookieSlice,
	}

	var out oidc.OidcTokens
	// FIXME: if we can adjust the request headers, we should forward the cookies in the request headers
	// We should use the GET endpoint and remove the POST endpoint once microcluster is updated to allow adjusting request headers
	err := c.Query(queryCtx, "POST", types.NoPrefix, targetURL, payload, &out)
	if err != nil {
		clientURL := c.URL()
		return oidc.OidcTokens{}, fmt.Errorf("Failed performing action on %q: %w", clientURL.String(), err)
	}

	return out, nil
}

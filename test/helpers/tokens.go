package helpers

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"time"

	"github.com/canonical/lxd/shared/api"

	"github.com/canonical/lxd-site-manager/internal/api/types"
)

// FindToken search for a token by site name.
func FindToken(env *Environment, siteName string) (types.ExternalSiteToken, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	unixClient, err := NewUnixHTTPClient(env.ControlSocketURL())
	if err != nil {
		return types.ExternalSiteToken{}, err
	}

	output := &[]types.ExternalSiteToken{}
	path := api.NewURL().Path("1.0", "external-site-join-token")
	err = unixClient.Query(ctx, http.MethodGet, path, nil, output, nil)
	if err != nil {
		return types.ExternalSiteToken{}, err
	}

	for _, token := range *output {
		if token.SiteName == siteName {
			return token, nil
		}
	}

	return types.ExternalSiteToken{}, nil
}

// CreateAndReturnSiteJoinToken creates a site join token and returns the decoded token data.
func CreateAndReturnSiteJoinToken(env *Environment, siteName string, expiry time.Time) (types.ExternalSiteTokenBody, error) {
	token, err := createSiteJoinToken(env, siteName, expiry)
	if err != nil {
		return types.ExternalSiteTokenBody{}, err
	}

	tokenData, err := decodeSiteJoinToken(token)
	if err != nil {
		return types.ExternalSiteTokenBody{}, err
	}

	return tokenData, nil
}

// decodeSiteJoinToken decodes a base 64 encoded external site join token.
func decodeSiteJoinToken(token string) (types.ExternalSiteTokenBody, error) {
	var tokenData types.ExternalSiteTokenBody
	decodedToken, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return types.ExternalSiteTokenBody{}, err
	}

	err = json.Unmarshal(decodedToken, &tokenData)
	if err != nil {
		return types.ExternalSiteTokenBody{}, err
	}

	return tokenData, nil
}

// createSiteJoinToken sets up a unix http client and sends a request to the site manager to create a site join token.
func createSiteJoinToken(env *Environment, siteName string, expiry time.Time) (token string, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	unixClient, err := NewUnixHTTPClient(env.ControlSocketURL())
	if err != nil {
		return "", err
	}

	input := types.ExternalSiteTokenPost{SiteName: siteName}
	if expiry != (time.Time{}) {
		input.Expiry = expiry
	}

	output := &types.ExternalSiteTokenPostResponse{}

	path := api.NewURL().Path("1.0", "external-site-join-token")
	err = unixClient.Query(ctx, http.MethodPost, path, input, output, nil)
	if err != nil {
		return "", err
	}

	return output.Token, nil
}

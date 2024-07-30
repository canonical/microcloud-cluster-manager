package helpers

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"time"

	"github.com/canonical/lxd/shared/api"

	"github.com/canonical/lxd-cluster-manager/internal/api/types"
)

// FindToken search for a token by remote cluster name.
func FindToken(env *Environment, remoteClusterName string) (types.RemoteClusterToken, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	unixClient, err := NewUnixHTTPClient(env.ControlSocketURL())
	if err != nil {
		return types.RemoteClusterToken{}, err
	}

	output := &[]types.RemoteClusterToken{}
	path := api.NewURL().Path("1.0", "remote-cluster-join-token")
	err = unixClient.Query(ctx, http.MethodGet, path, nil, output, nil)
	if err != nil {
		return types.RemoteClusterToken{}, err
	}

	for _, token := range *output {
		if token.ClusterName == remoteClusterName {
			return token, nil
		}
	}

	return types.RemoteClusterToken{}, nil
}

// CreateAndReturnRemoteClusterJoinToken creates a remote cluster join token and returns the decoded token data.
func CreateAndReturnRemoteClusterJoinToken(env *Environment, remoteClusterName string, expiry time.Time) (types.RemoteClusterTokenBody, error) {
	token, err := createRemoteClusterJoinToken(env, remoteClusterName, expiry)
	if err != nil {
		return types.RemoteClusterTokenBody{}, err
	}

	tokenData, err := decodeRemoteClusterJoinToken(token)
	if err != nil {
		return types.RemoteClusterTokenBody{}, err
	}

	return tokenData, nil
}

// decodeRemoteClusterJoinToken decodes a base 64 encoded external remote cluster join token.
func decodeRemoteClusterJoinToken(token string) (types.RemoteClusterTokenBody, error) {
	var tokenData types.RemoteClusterTokenBody
	decodedToken, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return types.RemoteClusterTokenBody{}, err
	}

	err = json.Unmarshal(decodedToken, &tokenData)
	if err != nil {
		return types.RemoteClusterTokenBody{}, err
	}

	return tokenData, nil
}

// createRemoteClusterJoinToken sets up a unix http client and sends a request to the Cluster Manager to create a remote cluster join token.
func createRemoteClusterJoinToken(env *Environment, remoteClusterName string, expiry time.Time) (token string, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	unixClient, err := NewUnixHTTPClient(env.ControlSocketURL())
	if err != nil {
		return "", err
	}

	input := types.RemoteClusterTokenPost{ClusterName: remoteClusterName}
	if expiry != (time.Time{}) {
		input.Expiry = expiry
	}

	output := &types.RemoteClusterTokenPostResponse{}

	path := api.NewURL().Path("1.0", "remote-cluster-join-token")
	err = unixClient.Query(ctx, http.MethodPost, path, input, output, nil)
	if err != nil {
		return "", err
	}

	return output.Token, nil
}

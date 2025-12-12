package helpers

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"time"

	"github.com/canonical/lxd/shared/api"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/api/models/v1"
)

// FindToken search for a token by remote cluster name.
func FindToken(env *Environment, remoteClusterName string) (models.RemoteClusterToken, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	headers, err := env.ManagementAPILoginHeaders()
	if err != nil {
		return models.RemoteClusterToken{}, err
	}

	certPublicKey, err := env.ManagementAPICert().PublicKeyX509()
	if err != nil {
		return models.RemoteClusterToken{}, err
	}

	tlsClient, err := NewTLSHTTPClient(api.URL{}, nil, certPublicKey, env.ManagementAPIHost())
	if err != nil {
		return models.RemoteClusterToken{}, err
	}

	output := &[]models.RemoteClusterToken{}
	path := api.NewURL().Scheme("https").Host(env.ManagementAPIHostPort()).Path("1.0", "remote-cluster-join-token")
	err = tlsClient.Query(ctx, http.MethodGet, path, nil, output, headers)
	if err != nil {
		return models.RemoteClusterToken{}, err
	}

	for _, token := range *output {
		if token.ClusterName == remoteClusterName {
			return token, nil
		}
	}

	return models.RemoteClusterToken{}, nil
}

// CreateAndReturnRemoteClusterJoinToken creates a remote cluster join token and returns the decoded token data.
func CreateAndReturnRemoteClusterJoinToken(env *Environment, remoteClusterName string, expiry time.Time) (models.RemoteClusterTokenBody, error) {
	token, err := createRemoteClusterJoinToken(env, remoteClusterName, expiry)
	if err != nil {
		return models.RemoteClusterTokenBody{}, err
	}

	tokenData, err := DecodeRemoteClusterJoinToken(token)
	if err != nil {
		return models.RemoteClusterTokenBody{}, err
	}

	return tokenData, nil
}

// DecodeRemoteClusterJoinToken decodes a base 64 encoded external remote cluster join token.
func DecodeRemoteClusterJoinToken(token string) (models.RemoteClusterTokenBody, error) {
	var tokenData models.RemoteClusterTokenBody
	decodedToken, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return models.RemoteClusterTokenBody{}, err
	}

	err = json.Unmarshal(decodedToken, &tokenData)
	if err != nil {
		return models.RemoteClusterTokenBody{}, err
	}

	return tokenData, nil
}

// createRemoteClusterJoinToken sets up a unix http client and sends a request to the Cluster Manager to create a remote cluster join token.
func createRemoteClusterJoinToken(env *Environment, remoteClusterName string, expiry time.Time) (token string, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	headers, err := env.ManagementAPILoginHeaders()
	if err != nil {
		return "", err
	}

	certPublicKey, err := env.ManagementAPICert().PublicKeyX509()
	if err != nil {
		return "", err
	}

	tlsClient, err := NewTLSHTTPClient(api.URL{}, nil, certPublicKey, env.ManagementAPIHost())
	if err != nil {
		return "", err
	}

	input := models.RemoteClusterTokenPost{ClusterName: remoteClusterName}
	if expiry != (time.Time{}) {
		input.Expiry = expiry
	}

	output := &models.RemoteClusterTokenPostResponse{}

	path := api.NewURL().Scheme("https").Host(env.ManagementAPIHostPort()).Path("1.0", "remote-cluster-join-token")
	err = tlsClient.Query(ctx, http.MethodPost, path, input, output, headers)
	if err != nil {
		return "", err
	}

	return output.Token, nil
}

// DeleteRemoteClusterJoinToken deletes a remote cluster join token by remote cluster name.
func DeleteRemoteClusterJoinToken(env *Environment, remoteClusterName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	certPublicKey, err := env.ManagementAPICert().PublicKeyX509()
	if err != nil {
		return err
	}

	tlsClient, err := NewTLSHTTPClient(api.URL{}, nil, certPublicKey, env.ManagementAPIHost())
	if err != nil {
		return err
	}

	path := api.NewURL().Scheme("https").Host(env.ManagementAPIHostPort()).Path("1.0", "remote-cluster-join-token", remoteClusterName)
	return tlsClient.Query(ctx, http.MethodDelete, path, nil, nil, nil)
}

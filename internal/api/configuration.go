package api

import (
	"embed"
	"encoding/json"
	"net/http"

	"github.com/canonical/lxd/lxd/response"
	"github.com/canonical/microcluster/v2/rest"
	microState "github.com/canonical/microcluster/v2/state"

	"github.com/canonical/lxd-cluster-manager/internal/state"
)

//go:embed metadata/configuration.json
var configJSON embed.FS

func metadataConfigurationCmd(s *state.ClusterManagerState) rest.Endpoint {
	return rest.Endpoint{
		Path: "metadata/configuration",
		Get: rest.EndpointAction{
			Handler:        metadataConfigurationGet,
			AllowUntrusted: true,
			AccessHandler:  oidcAuthHandler(s),
		},
	}
}

func metadataConfigurationGet(_ microState.State, r *http.Request) response.Response {
	file, err := configJSON.ReadFile("metadata/configuration.json")
	if err != nil {
		return response.SmartError(err)
	}

	var data map[string]any
	err = json.Unmarshal(file, &data)
	if err != nil {
		return response.SmartError(err)
	}

	return response.SyncResponse(true, data)
}

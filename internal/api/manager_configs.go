package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/canonical/lxd/lxd/db/query"
	"github.com/canonical/lxd/lxd/response"
	"github.com/canonical/lxd/lxd/util"
	"github.com/canonical/microcluster/client"
	"github.com/canonical/microcluster/rest"
	microState "github.com/canonical/microcluster/state"

	"github.com/canonical/lxd-cluster-manager/internal/api/types"
	clusterManagerClient "github.com/canonical/lxd-cluster-manager/internal/client"
	"github.com/canonical/lxd-cluster-manager/internal/database"
	"github.com/canonical/lxd-cluster-manager/internal/oidc"
	"github.com/canonical/lxd-cluster-manager/internal/state"
)

var managerConfigsCmd = func(s *state.ClusterManagerState) rest.Endpoint {
	return rest.Endpoint{
		Path: "config",
		Patch: rest.EndpointAction{
			Handler:        managerConfigPatch(s),
			AllowUntrusted: true,
			AccessHandler:  authHandler(s),
		},
		Get: rest.EndpointAction{
			Handler:        managerConfigsGet,
			AllowUntrusted: true,
			AccessHandler:  authHandler(s),
		},
	}
}

// partially update manager configs, replace configs only if they exist in payload.
func managerConfigPatch(clusterManagerState *state.ClusterManagerState) types.EndpointHandler {
	return func(clusterState microState.State, r *http.Request) response.Response {
		var payload types.ManagerConfigs

		err := json.NewDecoder(r.Body).Decode(&payload)
		if err != nil {
			return response.BadRequest(err)
		}

		if payload.Config == nil {
			return response.BadRequest(fmt.Errorf("missing config key"))
		}

		err = types.ValidateManagerConfigKeys(payload.Config)
		if err != nil {
			return response.BadRequest(err)
		}

		// If the request is not a notification, we need to notify the other cluster members about the same config update
		if !client.IsNotification(r) {
			cluster, err := clusterState.Cluster(true)
			if err != nil {
				return response.SmartError(err)
			}

			err = cluster.Query(r.Context(), true, func(ctx context.Context, c *client.Client) error {
				err := clusterManagerClient.ManagerConfigsPatchCmd(ctx, c, &payload)
				if err != nil {
					clientURL := c.URL()
					return fmt.Errorf("Failed to notify cluster member with address %q: %w", clientURL.String(), err)
				}

				return nil
			})

			if err != nil {
				return response.SmartError(err)
			}
		}

		updatedConfigs := make(map[string]string)
		err = clusterState.Database().Transaction(r.Context(), func(ctx context.Context, tx *sql.Tx) error {
			err := query.UpdateConfig(tx, "manager_configs", payload.Config)
			if err != nil {
				return err
			}

			dbConfigs, err := database.GetManagerConfig(ctx, tx)
			if err != nil {
				return err
			}

			for _, c := range dbConfigs {
				updatedConfigs[c.Key] = c.Value
			}

			return nil
		})

		if err != nil {
			return response.SmartError(err)
		}

		if OIDCConfigsChanged(payload.Config) {
			err = UpdateDaemonOIDCConfig(updatedConfigs, clusterManagerState, clusterState)
			if err != nil {
				return response.SmartError(err)
			}
		}

		return response.EmptySyncResponse
	}
}

func managerConfigsGet(s microState.State, r *http.Request) response.Response {
	var dbConfigs []database.ManagerConfig
	err := s.Database().Transaction(r.Context(), func(ctx context.Context, tx *sql.Tx) error {
		var err error
		dbConfigs, err = database.GetManagerConfig(ctx, tx)
		return err
	})

	if err != nil {
		return response.SmartError(err)
	}

	return response.SyncResponse(true, toManagerConfigsAPI(dbConfigs))
}

func toManagerConfigsAPI(dbConfigs []database.ManagerConfig) types.ManagerConfigs {
	configs := types.ManagerConfigs{
		Config: map[string]string{},
	}

	for _, c := range dbConfigs {
		configs.Config[c.Key] = c.Value
	}

	return configs
}

// OIDCConfigsChanged checks if the OIDC configs have changed.
func OIDCConfigsChanged(newConfigs map[string]string) bool {
	for key := range newConfigs {
		if key == "oidc.client.id" || key == "oidc.issuer" || key == "oidc.audience" {
			return true
		}
	}

	return false
}

// UpdateDaemonOIDCConfig updates the daemon OIDC configs.
func UpdateDaemonOIDCConfig(
	configs map[string]string,
	clusterManagerState *state.ClusterManagerState,
	clusterState microState.State,
) error {
	OIDCIssuer, issuerOk := configs["oidc.issuer"]
	OIDCClientID, clientIDOk := configs["oidc.client.id"]
	OIDCAudience, audienceOk := configs["oidc.audience"]

	// If any of the OIDC config is missing, we don't update the daemon oidc configs.
	if !issuerOk || !clientIDOk || !audienceOk {
		return nil
	}

	httpClientFunc := func() (*http.Client, error) {
		return util.HTTPClient("", http.ProxyFromEnvironment)
	}

	oidcVerifier, err := oidc.NewVerifier(
		OIDCIssuer,
		OIDCClientID,
		OIDCAudience,
		clusterState.ClusterCert,
		httpClientFunc,
	)

	if err != nil {
		return err
	}

	clusterManagerState.SetOIDCVerifier(oidcVerifier)

	return nil
}

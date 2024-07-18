package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/canonical/lxd/lxd/response"
	"github.com/canonical/lxd/shared/logger"
	"github.com/canonical/lxd/shared/revert"
	microClient "github.com/canonical/microcluster/client"
	"github.com/canonical/microcluster/rest"
	microTypes "github.com/canonical/microcluster/rest/types"
	clusterState "github.com/canonical/microcluster/state"
	"github.com/gorilla/mux"

	"github.com/canonical/lxd-site-manager/internal/api/types"
	"github.com/canonical/lxd-site-manager/internal/client"
	"github.com/canonical/lxd-site-manager/internal/database"
	"github.com/canonical/lxd-site-manager/internal/state"
)

func memberConfigCmd(s *state.SiteManagerState) rest.Endpoint {
	return rest.Endpoint{
		Path: "member/{name}/config",
		Patch: rest.EndpointAction{
			Handler:        memberConfigPatch(s),
			AllowUntrusted: true,
			AccessHandler:  authHandler(s),
		},
		Get: rest.EndpointAction{
			Handler:        memberConfigGet,
			AllowUntrusted: true,
			AccessHandler:  authHandler(s),
		},
	}
}

func memberConfigsCmd(s *state.SiteManagerState) rest.Endpoint {
	return rest.Endpoint{
		Path: "member/config",
		Get: rest.EndpointAction{
			Handler:        memberConfigsGet,
			AllowUntrusted: true,
			AccessHandler:  authHandler(s),
		},
	}
}

// update existing member configs.
func memberConfigPatch(siteManagerState *state.SiteManagerState) types.EndpointHandler {
	return func(clusterState clusterState.State, r *http.Request) response.Response {
		memberName, err := url.PathUnescape(mux.Vars(r)["name"])
		if err != nil {
			return response.BadRequest(err)
		}

		var payload types.MemberConfigPatch
		err = json.NewDecoder(r.Body).Decode(&payload)
		if err != nil {
			return response.BadRequest(err)
		}

		HTTPSAddress, hasHTTPSAddress := payload.Config[types.HTTPSAddress]
		externalAddress, hasExternalAddress := payload.Config[types.ExternalAddress]

		if !hasHTTPSAddress && !hasExternalAddress {
			return response.BadRequest(fmt.Errorf("no fields provided to update"))
		}

		if hasExternalAddress && externalAddress != "" {
			_, err = microTypes.ParseAddrPort(externalAddress)
			if err != nil {
				return response.BadRequest(fmt.Errorf("invalid external_address for member %q: %w", memberName, err))
			}
		}

		reverter := revert.New()
		defer reverter.Fail()

		queryClient, err := getClientByName(clusterState, siteManagerState, memberName)
		if err != nil {
			return response.InternalError(fmt.Errorf("failed to get client for member %q: %w", memberName, err))
		}

		if hasHTTPSAddress {
			newAddress, err := microTypes.ParseAddrPort(HTTPSAddress)
			if err != nil {
				return response.BadRequest(fmt.Errorf("invalid https_address for member %q: %w", memberName, err))
			}

			// the control listener address is stored in a member's local state directory
			// we need to update the control listener address, we need to forward the request to the relevant member and let the update happen there
			if memberName != clusterState.Name() {
				queryClient.SetClusterNotification()
				err = client.MemberConfigPatchCmd(r.Context(), queryClient, memberName, &payload)
				if err != nil {
					return response.InternalError(fmt.Errorf("failed to update member %q config: %w", memberName, err))
				}

				return response.EmptySyncResponse
			}

			// update the control listener address for local member configs
			newServerConfig := make(map[string]microTypes.ServerConfig)
			existingServerConfig, err := client.GetDaemonServerConfigs(r.Context(), queryClient)
			if err != nil {
				return response.InternalError(fmt.Errorf("failed to get local member %q config: %w", memberName, err))
			}

			for name, config := range existingServerConfig {
				if name == string(ControlListener) {
					config.Address = newAddress
				}

				newServerConfig[name] = config
			}

			err = siteManagerState.MicroCluster.UpdateServers(r.Context(), newServerConfig)
			if err != nil {
				return response.InternalError(fmt.Errorf("failed to update local member %q config: %w", memberName, err))
			}

			// in case if the dqlite transaction fails to update the member config, we need to revert the control listener address update
			// this will keep daemon local configs in sync with what's stored in dqlite
			reverter.Add(func() {
				err := siteManagerState.MicroCluster.UpdateServers(r.Context(), existingServerConfig)
				if err != nil {
					logger.Warn("Failed to revert control listener address update, data may be inconsistent")
				}

				logger.Warn("Reverted control listener address update")
			})
		}

		err = clusterState.Database().Transaction(r.Context(), func(ctx context.Context, tx *sql.Tx) error {
			// get existing member config entry and use it as a base
			filter := database.ManagerMemberConfigFilter{
				Target: &memberName,
			}

			dbConfigs, err := database.GetManagerMemberConfig(ctx, tx, filter)
			if err != nil {
				return err
			}

			// if no external address is provided, keep the existing one
			if !hasExternalAddress {
				externalAddress = dbConfigs[0].ExternalAddress
			}

			serverConfigs, err := client.GetDaemonServerConfigs(r.Context(), queryClient)
			if err != nil {
				return fmt.Errorf("failed to get local member %q config: %w", memberName, err)
			}

			controlListenerConfig, ok := serverConfigs[string(ControlListener)]
			if !ok {
				return fmt.Errorf("control listener config not found")
			}

			httpsAddress := controlListenerConfig.Address.String()

			// It is expected a member config entry was created for every member during initialisation
			return database.UpdateManagerMemberConfig(ctx, tx, memberName, database.ManagerMemberConfig{
				Target:          memberName,
				HTTPSAddress:    httpsAddress,
				ExternalAddress: externalAddress,
			})
		})

		if err != nil {
			return response.SmartError(err)
		}

		reverter.Success()

		return response.EmptySyncResponse
	}
}

func memberConfigGet(s clusterState.State, r *http.Request) response.Response {
	memberName, err := url.PathUnescape(mux.Vars(r)["name"])
	if err != nil {
		return response.BadRequest(err)
	}

	var dbConfigs []database.ManagerMemberConfig
	err = s.Database().Transaction(r.Context(), func(ctx context.Context, tx *sql.Tx) error {
		var err error
		filter := database.ManagerMemberConfigFilter{
			Target: &memberName,
		}

		dbConfigs, err = database.GetManagerMemberConfig(ctx, tx, filter)
		return err
	})

	if err != nil {
		return response.SmartError(err)
	}

	if len(dbConfigs) == 0 {
		return response.NotFound(fmt.Errorf("Member not found"))
	}

	return response.SyncResponse(true, toMemberConfigsAPI(dbConfigs)[0])
}

func memberConfigsGet(s clusterState.State, r *http.Request) response.Response {
	var dbConfigs []database.ManagerMemberConfig
	err := s.Database().Transaction(r.Context(), func(ctx context.Context, tx *sql.Tx) error {
		var err error
		dbConfigs, err = database.GetManagerMemberConfig(ctx, tx)
		return err
	})

	if err != nil {
		return response.SmartError(err)
	}

	return response.SyncResponse(true, toMemberConfigsAPI(dbConfigs))
}

func toMemberConfigsAPI(dbConfigs []database.ManagerMemberConfig) []types.MemberConfig {
	var memberConfigs []types.MemberConfig
	for _, c := range dbConfigs {
		memberConfigs = append(memberConfigs, types.MemberConfig{
			Target: c.Target,
			MemberConfigPatch: types.MemberConfigPatch{
				Config: map[types.MemberConfigKey]string{
					types.HTTPSAddress:    c.HTTPSAddress,
					types.ExternalAddress: c.ExternalAddress,
				},
			},
		})
	}

	return memberConfigs
}

func getClientByName(clusterState clusterState.State, siteManagerState *state.SiteManagerState, name string) (*microClient.Client, error) {
	if clusterState.Name() == name {
		localClient, err := siteManagerState.MicroCluster.LocalClient()
		if err != nil {
			return nil, fmt.Errorf("failed to get local client: %w", err)
		}

		return localClient, nil
	}

	remotes := clusterState.Remotes().RemotesByName()
	targetRemote, ok := remotes[name]
	if !ok {
		return nil, fmt.Errorf("member %q not found", name)
	}

	client, err := siteManagerState.MicroCluster.RemoteClient(targetRemote.Address.String())
	if err != nil {
		return nil, fmt.Errorf("failed to get remote client for member: %s", name)
	}

	return client, nil
}

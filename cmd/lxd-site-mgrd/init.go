package main

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/canonical/microcluster/microcluster"
	"github.com/canonical/microcluster/rest/types"
	"github.com/canonical/microcluster/state"

	"github.com/canonical/lxd-site-manager/internal/api"
	"github.com/canonical/lxd-site-manager/internal/database"
)

// InitialiseControlListener is a hook that initialises the control listener for the daemon.
// It will update the local daemon server configs with the control listener address and start the control listener.
// It will also create a member config entry in dqlite with he control listener address.
func InitialiseControlListener(ctx context.Context, s state.State, m *microcluster.MicroCluster, initConfig map[string]string) error {
	controlListenerAddress, ok := initConfig[string(api.ControlListener)]
	if !ok {
		return fmt.Errorf("control listener address not provided")
	}

	address, err := types.ParseAddrPort(controlListenerAddress)
	if err != nil {
		return err
	}

	serverConfig := map[string]types.ServerConfig{
		string(api.ControlListener): {
			Address: address,
		},
	}

	err = m.UpdateServers(ctx, serverConfig)
	if err != nil {
		return fmt.Errorf("failed to update initialise control listener: %w", err)
	}

	return s.Database().Transaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
		var err error
		_, err = database.CreateManagerMemberConfig(ctx, tx, database.ManagerMemberConfig{
			Target:          s.Name(),
			HTTPSAddress:    controlListenerAddress,
			ExternalAddress: "",
		})

		return err
	})
}

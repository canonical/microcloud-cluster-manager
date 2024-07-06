package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/canonical/lxd/lxd/db/query"
	"github.com/canonical/lxd/lxd/response"
	"github.com/canonical/microcluster/rest"
	"github.com/canonical/microcluster/state"

	"github.com/canonical/lxd-site-manager/internal/api/types"
	"github.com/canonical/lxd-site-manager/internal/database"
)

var managerConfigsCmd = rest.Endpoint{
	Path:  "config",
	Patch: rest.EndpointAction{Handler: managerConfigPatch, AllowUntrusted: true},
	Get:   rest.EndpointAction{Handler: managerConfigsGet, AllowUntrusted: true},
}

// partially update manager configs, replace configs only if they exist in payload.
func managerConfigPatch(s *state.State, r *http.Request) response.Response {
	var payload types.ManagerConfigs

	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		return response.BadRequest(err)
	}

	if payload.Config == nil {
		return response.BadRequest(fmt.Errorf("missing config key"))
	}

	err = types.ValidateConfigKeys(payload.Config)
	if err != nil {
		return response.BadRequest(err)
	}

	err = s.Database.Transaction(r.Context(), func(ctx context.Context, tx *sql.Tx) error {
		err := query.UpdateConfig(tx, "manager_configs", payload.Config)
		return err
	})

	if err != nil {
		return response.SmartError(err)
	}

	return response.EmptySyncResponse
}

func managerConfigsGet(s *state.State, r *http.Request) response.Response {
	var dbConfigs []database.ManagerConfig
	err := s.Database.Transaction(r.Context(), func(ctx context.Context, tx *sql.Tx) error {
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

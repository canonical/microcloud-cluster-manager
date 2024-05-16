package api

import (
	"context"
	"database/sql"
	"net/http"
	"net/url"

	"github.com/canonical/lxd/lxd/response"
	"github.com/canonical/microcluster/rest"
	"github.com/canonical/microcluster/state"
	"github.com/gorilla/mux"

	"github.com/canonical/lxd-site-manager/api/types"
	"github.com/canonical/lxd-site-manager/database"
)

// This is an example extended endpoint on the /1.0 endpoint, reachable at /1.0/extended.
var sitesCmd = rest.Endpoint{
	Path: "sites",

	Get: rest.EndpointAction{Handler: sitesGet, AllowUntrusted: true},
}

var siteCmd = rest.Endpoint{
	Path: "sites/{siteName}",
	Get:  rest.EndpointAction{Handler: siteGet, AllowUntrusted: true},
}

func sitesGet(s *state.State, r *http.Request) response.Response {
	var dbSites []database.Site
	err := s.Database.Transaction(r.Context(), func(ctx context.Context, tx *sql.Tx) error {
		var err error
		dbSites, err = database.GetSites(ctx, tx)
		return err
	})
	if err != nil {
		return response.SmartError(err)
	}

	apiSites := make([]types.Site, 0, len(dbSites))
	for _, dbSite := range dbSites {
		apiSites = append(apiSites, types.Site{
			Name:      dbSite.Name,
			Addresses: dbSite.Addresses,
			Status:    dbSite.Status,
		})
	}

	return response.SyncResponse(true, apiSites)
}

func siteGet(s *state.State, r *http.Request) response.Response {
	siteName, err := url.PathUnescape(mux.Vars(r)["siteName"])
	if err != nil {
		return response.SmartError(err)
	}

	var dbSite *database.Site
	err = s.Database.Transaction(r.Context(), func(ctx context.Context, tx *sql.Tx) error {
		var err error
		dbSite, err = database.GetSite(ctx, tx, siteName)
		return err
	})
	if err != nil {
		return response.SmartError(err)
	}

	return response.SyncResponse(true, types.Site{
		Name:      dbSite.Name,
		Addresses: dbSite.Addresses,
		Status:    dbSite.Status,
	})
}

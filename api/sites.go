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
	"github.com/canonical/microcluster/client"
	"github.com/canonical/microcluster/rest"
	"github.com/canonical/microcluster/state"
	"github.com/gorilla/mux"

	"github.com/canonical/lxd-site-manager/api/types"
	localClient "github.com/canonical/lxd-site-manager/client"
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

var sitePostCmd = rest.Endpoint{
	Path: "sites",
	Post: rest.EndpointAction{Handler: sitePost, AllowUntrusted: true},
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

func sitePost(state *state.State, r *http.Request) response.Response {
	// Check the user agent header to check if we are the notifying cluster member.
	if !client.IsNotification(r) {
		// Get a collection of clients every other cluster member, with the notification user-agent set.
		cluster, err := state.Cluster(true)
		if err != nil {
			return response.SmartError(fmt.Errorf("failed to get a client for every cluster member: %w", err))
		}

		var site types.Site
		err = json.NewDecoder(r.Body).Decode(&site)
		if err != nil {
			logger.Errorf("Failed decoding site data: %v", err)
			return response.InternalError(err)
		}

		messages := make([]string, 0, len(cluster))
		err = cluster.Query(state.Context, true, func(ctx context.Context, c *client.Client) error {
			if err != nil {
				return fmt.Errorf("failed to parse addr:port of listen address %q: %w", state.Address().URL.Host, err)
			}

			outMessage, err := localClient.PostSite(ctx, c, &site)
			if err != nil {
				clientURL := c.URL()
				return fmt.Errorf("failed to POST to cluster member with address %q: %w", clientURL.String(), err)
			}

			messages = append(messages, outMessage)

			return nil
		})

		if err != nil {
			return response.SmartError(err)
		}

		// Having received the result from all forwarded requests, compile them as a string and return.
		var outMsg string
		for _, message := range messages {
			outMsg = outMsg + message + "\n"
		}

		return response.SyncResponse(true, outMsg)
	}

	// Decode the POST body using our defined ExtendedType.
	var data types.Site
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		return response.SmartError(err)
	}

	err = state.Database.Transaction(state.Context, func(ctx context.Context, tx *sql.Tx) error {
		_, err = database.CreateSite(ctx, tx, &database.Site{Name: data.Name, Status: data.Status})
		if err != nil {
			return fmt.Errorf("failed to save site: %w", err)
		}
		return nil
	})

	if err != nil {
		return response.SmartError(err)
	}

	message := fmt.Sprintf("Site created %s", data.Name)

	return response.SyncResponse(true, message)
}

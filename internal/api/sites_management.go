package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/canonical/lxd/lxd/response"
	"github.com/canonical/microcluster/rest"
	microState "github.com/canonical/microcluster/state"
	"github.com/gorilla/mux"

	"github.com/canonical/lxd-site-manager/internal/api/types"
	"github.com/canonical/lxd-site-manager/internal/database"
	"github.com/canonical/lxd-site-manager/internal/state"
)

func sitesCmd(s *state.SiteManagerState) rest.Endpoint {
	return rest.Endpoint{
		Path: "sites",
		Get: rest.EndpointAction{
			Handler:        sitesGet,
			AllowUntrusted: true,
			AccessHandler:  authHandler(s),
		},
	}
}

func siteCmd(s *state.SiteManagerState) rest.Endpoint {
	return rest.Endpoint{
		Path: "sites/{siteName}",
		Get: rest.EndpointAction{
			Handler:        siteGet,
			AllowUntrusted: true,
			AccessHandler:  authHandler(s),
		},
		Delete: rest.EndpointAction{
			Handler:        siteDelete,
			AllowUntrusted: true,
			AccessHandler:  authHandler(s),
		},
		Patch: rest.EndpointAction{
			Handler:        sitePatch,
			AllowUntrusted: true,
			AccessHandler:  authHandler(s),
		},
	}
}

func sitesGet(s microState.State, r *http.Request) response.Response {
	var dbSiteDetails []database.CoreSiteWithDetails

	err := s.Database().Transaction(r.Context(), func(ctx context.Context, tx *sql.Tx) error {
		var err error
		dbSiteDetails, err = database.GetCoreSitesWithDetails(ctx, tx)
		return err
	})

	if err != nil {
		return response.SmartError(err)
	}

	result, err := toSitesAPI(dbSiteDetails)
	if err != nil {
		return response.InternalError(err)
	}

	return response.SyncResponse(true, result)
}

func siteGet(s microState.State, r *http.Request) response.Response {
	siteName, err := url.PathUnescape(mux.Vars(r)["siteName"])
	if err != nil {
		return response.SmartError(err)
	}

	var dbSiteDetails []database.CoreSiteWithDetails
	err = s.Database().Transaction(r.Context(), func(ctx context.Context, tx *sql.Tx) error {
		var err error
		dbSiteDetails, err = database.GetCoreSiteWithDetailBySiteName(ctx, tx, siteName)
		return err
	})

	if err != nil {
		return response.SmartError(err)
	}

	if len(dbSiteDetails) == 0 {
		return response.NotFound(fmt.Errorf("Site not found"))
	}

	result, err := toSitesAPI(dbSiteDetails)
	if err != nil {
		return response.InternalError(err)
	}

	return response.SyncResponse(true, result[0])
}

func siteDelete(s microState.State, r *http.Request) response.Response {
	siteName, err := url.PathUnescape(mux.Vars(r)["siteName"])
	if err != nil {
		return response.SmartError(err)
	}

	err = s.Database().Transaction(r.Context(), func(ctx context.Context, tx *sql.Tx) error {
		return database.DeleteCoreSite(ctx, tx, siteName)
	})

	if err != nil {
		return response.SmartError(err)
	}

	return response.EmptySyncResponse
}

func sitePatch(s microState.State, r *http.Request) response.Response {
	siteName, err := url.PathUnescape(mux.Vars(r)["siteName"])
	if err != nil {
		return response.SmartError(err)
	}

	var payload types.SitePatch
	err = json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		return response.BadRequest(err)
	}

	if payload.Status != "" {
		if payload.Status != types.PENDING_APPROVAL && payload.Status != types.ACTIVE {
			return response.BadRequest(fmt.Errorf("Invalid status"))
		}
	}

	err = s.Database().Transaction(r.Context(), func(ctx context.Context, tx *sql.Tx) error {
		siteID, err := database.GetCoreSiteID(ctx, tx, siteName)
		if err != nil {
			return err
		}

		existingSiteDetail, err := database.GetSiteDetail(ctx, tx, siteID)
		if err != nil {
			return err
		}

		newSiteDetail := existingSiteDetail
		if payload.Status != "" {
			if existingSiteDetail.Status == string(types.PENDING_APPROVAL) && payload.Status == types.ACTIVE {
				newSiteDetail.JoinedAt = time.Now()
			}
			newSiteDetail.Status = string(payload.Status)
		}

		newSiteDetail.UpdatedAt = time.Now()

		return database.UpdateSiteDetail(ctx, tx, siteID, *newSiteDetail)
	})

	if err != nil {
		return response.SmartError(err)
	}

	return response.EmptySyncResponse
}

func toSitesAPI(dbEntries []database.CoreSiteWithDetails) ([]types.Site, error) {
	// generate lookup for site details
	var sites []types.Site
	for _, e := range dbEntries {
		var ms []types.StatusDistribution
		err := json.Unmarshal([]byte(e.MemberStatuses), &ms)
		if err != nil {
			return nil, fmt.Errorf("Failed to unmarshal member statuses: %w", err)
		}

		var is []types.StatusDistribution
		err = json.Unmarshal([]byte(e.InstanceStatuses), &is)
		if err != nil {
			return nil, fmt.Errorf("Failed to unmarshal instance statuses: %w", err)
		}

		sites = append(sites, types.Site{
			Name:               e.Name,
			SiteCertificate:    e.SiteCertificate,
			Status:             e.Status,
			CPUTotalCount:      e.CPUTotalCount,
			CPULoad1:           e.CPULoad1,
			CPULoad5:           e.CPULoad5,
			CPULoad15:          e.CPULoad15,
			MemoryTotalAmount:  e.MemoryTotalAmount,
			MemoryUsage:        e.MemoryUsage,
			DiskTotalSize:      e.DiskTotalSize,
			DiskUsage:          e.DiskUsage,
			MemberCount:        e.MemberCount,
			MemberStatuses:     ms,
			InstanceCount:      e.InstanceCount,
			InstanceStatuses:   is,
			JoinedAt:           e.SiteJoinedAt,
			CreatedAt:          e.SiteCreatedAt,
			LastStatusUpdateAt: e.SiteUpdatedAt,
		})
	}

	return sites, nil
}

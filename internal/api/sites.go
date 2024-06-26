package api

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/url"

	"github.com/canonical/lxd/lxd/response"
	"github.com/canonical/microcluster/rest"
	"github.com/canonical/microcluster/state"
	"github.com/gorilla/mux"

	"github.com/canonical/lxd-site-manager/internal/api/types"
	"github.com/canonical/lxd-site-manager/internal/database"
)

// This is an example extended endpoint on the /1.0 endpoint, reachable at /1.0/extended.
var sitesCmd = rest.Endpoint{
	Path: "sites",
	Get:  rest.EndpointAction{Handler: sitesGet, AllowUntrusted: true},
}

var siteCmd = rest.Endpoint{
	Path: "sites/{siteName}",
	Get:  rest.EndpointAction{Handler: siteGet, AllowUntrusted: true},
}

func sitesGet(s *state.State, r *http.Request) response.Response {
	var dbMemberStatusesForAllSites []database.MemberStatusWithSiteInfo

	err := s.Database.Transaction(r.Context(), func(ctx context.Context, tx *sql.Tx) error {
		var err error
		dbMemberStatusesForAllSites, err = database.GetMemberStatusesWithSiteInfo(ctx, tx)
		return err
	})

	if err != nil {
		return response.SmartError(err)
	}

	return response.SyncResponse(true, toSitesAPI(dbMemberStatusesForAllSites))
}

func siteGet(s *state.State, r *http.Request) response.Response {
	siteName, err := url.PathUnescape(mux.Vars(r)["siteName"])
	if err != nil {
		return response.SmartError(err)
	}

	var dbSiteMemberStatuses []database.MemberStatusWithSiteInfo
	err = s.Database.Transaction(r.Context(), func(ctx context.Context, tx *sql.Tx) error {
		var err error
		dbSiteMemberStatuses, err = database.GetMemberStatusesWithSiteInfoBySiteName(ctx, tx, siteName)
		return err
	})

	if err != nil {
		return response.SmartError(err)
	}

	if len(dbSiteMemberStatuses) == 0 {
		return response.NotFound(fmt.Errorf("Site not found"))
	}

	return response.SyncResponse(true, toSitesAPI(dbSiteMemberStatuses)[0])
}

func toSitesAPI(dbEntries []database.MemberStatusWithSiteInfo) []types.Site {
	// generate lookup for site details
	siteDetails := make(map[int]*types.Site)
	for _, e := range dbEntries {
		if _, ok := siteDetails[e.ID]; !ok {
			siteDetails[e.ID] = &types.Site{
				Name:               e.Name,
				SiteCertificate:    e.SiteCertificate,
				Status:             e.Status,
				InstanceCount:      e.InstanceCount,
				InstanceStatuses:   e.InstanceStatuses,
				JoinedAt:           e.SiteJoinedAt,
				CreatedAt:          e.SiteCreatedAt,
				LastStatusUpdateAt: e.SiteUpdatedAt,
				MemberStatuses:     []types.MemberStatus{},
			}
		}

		siteDetails[e.ID].MemberStatuses = append(
			siteDetails[e.ID].MemberStatuses,
			types.MemberStatus{
				Address:      e.Address,
				Architecture: e.Architecture,
				Role:         e.Role,
				UsageCPU:     e.UsageCPU,
				UsageMemory:  e.UsageMemory,
				UsageDisk:    e.UsageDisk,
				Status:       e.MemberStatus,
			},
		)
	}

	var sites []types.Site
	for _, s := range siteDetails {
		sites = append(sites, *s)
	}

	return sites
}

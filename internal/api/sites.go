package api

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/canonical/lxd/lxd/response"
	"github.com/canonical/microcluster/rest"
	"github.com/canonical/microcluster/state"
	"github.com/gorilla/mux"

	"github.com/canonical/lxd-site-manager/internal/api/types"
	"github.com/canonical/lxd-site-manager/internal/database"
)

var sitesCmd = rest.Endpoint{
	Path: "sites",
	Get:  rest.EndpointAction{Handler: sitesGet, AllowUntrusted: true},
	// FIXME: this endpoint will be a resource for the control listener
	Post: rest.EndpointAction{Handler: sitesPost, AllowUntrusted: true},
}

var siteCmd = rest.Endpoint{
	Path:   "sites/{siteName}",
	Get:    rest.EndpointAction{Handler: siteGet, AllowUntrusted: true},
	Delete: rest.EndpointAction{Handler: siteDelete, AllowUntrusted: true},
}

func sitesGet(s *state.State, r *http.Request) response.Response {
	var dbSiteDetails []database.CoreSiteWithDetails

	err := s.Database.Transaction(r.Context(), func(ctx context.Context, tx *sql.Tx) error {
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

func siteGet(s *state.State, r *http.Request) response.Response {
	siteName, err := url.PathUnescape(mux.Vars(r)["siteName"])
	if err != nil {
		return response.SmartError(err)
	}

	var dbSiteDetails []database.CoreSiteWithDetails
	err = s.Database.Transaction(r.Context(), func(ctx context.Context, tx *sql.Tx) error {
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

func siteDelete(s *state.State, r *http.Request) response.Response {
	siteName, err := url.PathUnescape(mux.Vars(r)["siteName"])
	if err != nil {
		return response.SmartError(err)
	}

	err = s.Database.Transaction(r.Context(), func(ctx context.Context, tx *sql.Tx) error {
		return database.DeleteCoreSite(ctx, tx, siteName)
	})

	if err != nil {
		return response.SmartError(err)
	}

	return response.EmptySyncResponse
}

func sitesPost(s *state.State, r *http.Request) response.Response {
	payload := types.SitePost{}

	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		return response.BadRequest(err)
	}

	if payload.SiteName == "" {
		return response.BadRequest(fmt.Errorf("site name is required"))
	}

	if payload.SiteCertificate == "" {
		return response.BadRequest(fmt.Errorf("site certificate is required"))
	}

	// get token secret for HMAC verification
	var token *database.CoreSiteToken
	err = s.Database.Transaction(r.Context(), func(ctx context.Context, tx *sql.Tx) error {
		var err error
		token, err = database.GetCoreSiteToken(ctx, tx, payload.SiteName)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return response.SmartError(err)
	}

	// check if token has expired
	if time.Now().After(token.Expiry) {
		return response.Forbidden(fmt.Errorf("token has expired"))
	}

	// verify HMAC
	hmacOK, err := verifyHMAC(payload, r, token.Secret)
	if err != nil || !hmacOK {
		return response.Forbidden(err)
	}

	// Create site entry and delete token in a single db transaction
	err = s.Database.Transaction(r.Context(), func(ctx context.Context, tx *sql.Tx) error {
		// create site entry
		siteID, err := database.CreateCoreSite(ctx, tx, database.CoreSite{
			Name:            payload.SiteName,
			SiteCertificate: payload.SiteCertificate,
		})

		if err != nil {
			return err
		}

		// create relevant site details
		_, err = database.CreateSiteDetail(ctx, tx, database.SiteDetail{
			Status:     string(types.PENDING_APPROVAL),
			CoreSiteID: siteID,
		})

		if err != nil {
			return err
		}

		// delete site token
		return database.DeleteCoreSiteToken(ctx, tx, payload.SiteName)
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
		var ms []types.Status
		err := json.Unmarshal([]byte(e.MemberStatuses), &ms)
		if err != nil {
			return nil, fmt.Errorf("Failed to unmarshal member statuses: %w", err)
		}

		var is []types.Status
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

func verifyHMAC(payload types.SitePost, r *http.Request, secret string) (bool, error) {
	reqBody, err := json.Marshal(payload)
	if err != nil {
		return false, fmt.Errorf("failed to marshal payload: %v", err)
	}

	sig := r.Header.Get("X-Site-Signature")
	if sig == "" {
		return false, fmt.Errorf("missing signature header")
	}

	decodedSig, err := base64.StdEncoding.DecodeString(sig)
	if err != nil {
		return false, fmt.Errorf("failed to decode signature: %v", err)
	}

	// recompute the HMAC
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(reqBody)
	expectedMac := mac.Sum(nil)

	return hmac.Equal(decodedSig, expectedMac), nil
}

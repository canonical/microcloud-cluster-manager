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
	"github.com/canonical/lxd/shared"
	"github.com/canonical/microcluster/rest"
	"github.com/canonical/microcluster/state"
	"github.com/gorilla/mux"

	"github.com/canonical/lxd-site-manager/internal/api/types"
	"github.com/canonical/lxd-site-manager/internal/database"
)

var externalSiteJoinTokensCmd = rest.Endpoint{
	Path: "external-site-join-token",
	Post: rest.EndpointAction{Handler: tokenPost, AllowUntrusted: true},
	Get:  rest.EndpointAction{Handler: tokenGet, AllowUntrusted: true},
}

var externalSiteJoinTokenCmd = rest.Endpoint{
	Path:   "external-site-join-token/{siteName}",
	Delete: rest.EndpointAction{Handler: tokenDelete, AllowUntrusted: true},
}

func tokenPost(s state.State, r *http.Request) response.Response {
	payload := types.ExternalSiteTokenPost{}

	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		return response.BadRequest(err)
	}

	if payload.Expiry == (time.Time{}) {
		return response.BadRequest(fmt.Errorf("token must have an expiry date"))
	}

	if payload.Expiry.Before(time.Now()) {
		return response.BadRequest(fmt.Errorf("expiry date must be in the future"))
	}

	if payload.SiteName == "" {
		return response.BadRequest(fmt.Errorf("site name is required"))
	}

	secret, err := shared.RandomCryptoString()
	if err != nil {
		return response.InternalError(err)
	}

	// store token details in the database
	err = s.Database().Transaction(r.Context(), func(ctx context.Context, tx *sql.Tx) error {
		var err error
		tokenData := database.CoreSiteToken{
			SiteName:  payload.SiteName,
			Secret:    secret,
			Expiry:    payload.Expiry,
			CreatedAt: time.Now(),
		}
		_, err = database.CreateCoreSiteToken(ctx, tx, tokenData)

		return err
	})

	if err != nil {
		return response.SmartError(err)
	}

	// create the token to be sent to LXD
	clusterCert, err := s.ClusterCert().PublicKeyX509()
	if err != nil {
		return response.InternalError(err)
	}

	memberAddresses, err := getSiteManagerAddresses(r.Context(), s)
	if err != nil {
		return response.InternalError(err)
	}

	token := types.ExternalSiteTokenBody{
		Secret:      secret,
		ExpiresAt:   payload.Expiry,
		Addresses:   memberAddresses,
		ServerName:  payload.SiteName,
		Fingerprint: shared.CertFingerprint(clusterCert),
	}

	encodedToken, err := token.Encode()
	if err != nil {
		return response.InternalError(err)
	}

	return response.SyncResponse(true, types.ExternalSiteTokenPostResponse{Token: encodedToken})
}

func tokenGet(s state.State, r *http.Request) response.Response {
	var tokens []database.CoreSiteToken
	err := s.Database().Transaction(r.Context(), func(ctx context.Context, tx *sql.Tx) error {
		var err error
		tokens, err = database.GetCoreSiteTokens(ctx, tx)
		return err
	})

	if err != nil {
		return response.SmartError(err)
	}

	var responseTokens []types.ExternalSiteToken
	for _, token := range tokens {
		responseTokens = append(responseTokens, types.ExternalSiteToken{
			Expiry:   token.Expiry,
			SiteName: token.SiteName,
			CreateAt: token.CreatedAt,
		})
	}

	return response.SyncResponse(true, responseTokens)
}

func tokenDelete(s state.State, r *http.Request) response.Response {
	siteName, err := url.PathUnescape(mux.Vars(r)["siteName"])
	if err != nil {
		return response.SmartError(err)
	}

	if siteName == "" {
		return response.BadRequest(fmt.Errorf("site name is required"))
	}

	err = s.Database().Transaction(r.Context(), func(ctx context.Context, tx *sql.Tx) error {
		return database.DeleteCoreSiteToken(ctx, tx, siteName)
	})

	if err != nil {
		return response.SmartError(err)
	}

	return response.SyncResponse(true, nil)
}

// getSiteManagerAddresses returns the addresses of the site managers that are online.
func getSiteManagerAddresses(ctx context.Context, s state.State) ([]string, error) {
	globalAddress, err := getGlobalAddress(s)
	if err != nil {
		return nil, err
	}

	if globalAddress != "" {
		return []string{globalAddress}, nil
	}

	memberConfigs, err := getMemberConfigs(s)
	if err != nil {
		return nil, err
	}

	return getLeaderPrioritisedAddresses(ctx, s, memberConfigs)
}

func getGlobalAddress(s state.State) (string, error) {
	var globalAddress string
	err := s.Database().Transaction(context.Background(), func(ctx context.Context, tx *sql.Tx) error {
		var err error
		managerConfigs, err := database.GetManagerConfig(ctx, tx)
		if err != nil {
			return err
		}

		globalAddress = toManagerConfigsAPI(managerConfigs).Config["global.address"]

		return nil
	})

	if err != nil {
		return "", err
	}

	return globalAddress, nil
}

func getMemberConfigs(s state.State) ([]database.ManagerMemberConfig, error) {
	var memberConfigs []database.ManagerMemberConfig
	err := s.Database().Transaction(context.Background(), func(ctx context.Context, tx *sql.Tx) error {
		var err error
		memberConfigs, err = database.GetManagerMemberConfig(ctx, tx)
		return err
	})

	if err != nil {
		return nil, err
	}

	return memberConfigs, nil
}

// getLeaderPrioritisedAddresses returns the addresses of site manager members (external if any is set) with the leader as the first address.
func getLeaderPrioritisedAddresses(ctx context.Context, s state.State, memberConfigs []database.ManagerMemberConfig) ([]string, error) {
	hasExternalAddresses := checkExternalAddresses(memberConfigs)
	leaderName, onlineMembers, err := getLeaderAndOnlineMembers(ctx, s)
	if err != nil {
		return nil, err
	}

	siteManagerAddresses := make([]string, 0, len(memberConfigs))
	nonOnlineMembers := make([]string, 0, len(memberConfigs))
	var leaderIndex int
	for i, memberConfig := range memberConfigs {
		address := memberConfig.HTTPSAddress

		// if any external addresses are set, use them
		if hasExternalAddresses {
			address = memberConfig.ExternalAddress
		}

		// it's possible that no external address is set for a member
		if address == "" {
			continue
		}

		_, ok := onlineMembers[memberConfig.Target]
		if !ok {
			nonOnlineMembers = append(nonOnlineMembers, address)
			continue
		}

		// if the member is the leader, capture its current index in the output list
		// We will use this index to swap the leader address to the first position in the list
		if memberConfig.Target == leaderName {
			leaderIndex = i
		}

		siteManagerAddresses = append(siteManagerAddresses, address)
	}

	if leaderIndex != 0 {
		siteManagerAddresses[0], siteManagerAddresses[leaderIndex] = siteManagerAddresses[leaderIndex], siteManagerAddresses[0]
	}

	return append(siteManagerAddresses, nonOnlineMembers...), nil
}

func checkExternalAddresses(memberConfigs []database.ManagerMemberConfig) bool {
	for _, memberConfig := range memberConfigs {
		if memberConfig.ExternalAddress != "" {
			return true
		}
	}

	return false
}

func getLeaderAndOnlineMembers(ctx context.Context, s state.State) (leaderName string, onlineMembers map[string]bool, err error) {
	leaderClient, err := s.Leader()
	if err != nil {
		return "", onlineMembers, err
	}

	clusterMembers, err := leaderClient.GetClusterMembers(ctx)
	if err != nil {
		return "", onlineMembers, err
	}

	leaderURL := leaderClient.URL()
	onlineMembers = make(map[string]bool)
	for _, clusterMember := range clusterMembers {
		if clusterMember.Status == "ONLINE" {
			onlineMembers[clusterMember.Name] = true
		}

		if clusterMember.Address.String() == leaderURL.URL.Host {
			leaderName = clusterMember.Name
		}
	}

	return leaderName, onlineMembers, nil
}

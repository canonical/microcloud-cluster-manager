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
	microState "github.com/canonical/microcluster/state"
	"github.com/gorilla/mux"

	"github.com/canonical/lxd-site-manager/internal/api/types"
	"github.com/canonical/lxd-site-manager/internal/database"
	"github.com/canonical/lxd-site-manager/internal/state"
)

func remoteClusterJoinTokensCmd(s *state.ClusterManagerState) rest.Endpoint {
	return rest.Endpoint{
		Path: "remote-cluster-join-token",
		Post: rest.EndpointAction{
			Handler:        tokenPost,
			AllowUntrusted: true,
			AccessHandler:  authHandler(s),
		},
		Get: rest.EndpointAction{Handler: tokenGet, AllowUntrusted: true},
	}
}

func remoteClusterJoinTokenCmd(s *state.ClusterManagerState) rest.Endpoint {
	return rest.Endpoint{
		Path: "remote-cluster-join-token/{remoteClusterName}",
		Delete: rest.EndpointAction{
			Handler:        tokenDelete,
			AllowUntrusted: true,
			AccessHandler:  authHandler(s),
		},
	}
}

func tokenPost(s microState.State, r *http.Request) response.Response {
	payload := types.RemoteClusterTokenPost{}

	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		return response.BadRequest(err)
	}

	// default expiry to 1 day if not set
	if payload.Expiry == (time.Time{}) {
		payload.Expiry = time.Now().Add(time.Hour * 24)
	}

	if payload.Expiry.Before(time.Now()) {
		return response.BadRequest(fmt.Errorf("expiry date must be in the future"))
	}

	if payload.ClusterName == "" {
		return response.BadRequest(fmt.Errorf("cluster name is required"))
	}

	secret, err := shared.RandomCryptoString()
	if err != nil {
		return response.InternalError(err)
	}

	// store token details in the database
	err = s.Database().Transaction(r.Context(), func(ctx context.Context, tx *sql.Tx) error {
		var err error
		tokenData := database.CoreRemoteClusterToken{
			ClusterName: payload.ClusterName,
			Secret:      secret,
			Expiry:      payload.Expiry,
			CreatedAt:   time.Now(),
		}
		_, err = database.CreateCoreRemoteClusterToken(ctx, tx, tokenData)

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

	memberAddresses, err := getClusterManagerAddresses(r.Context(), s)
	if err != nil {
		return response.InternalError(err)
	}

	token := types.RemoteClusterTokenBody{
		Secret:      secret,
		ExpiresAt:   payload.Expiry,
		Addresses:   memberAddresses,
		ServerName:  payload.ClusterName,
		Fingerprint: shared.CertFingerprint(clusterCert),
	}

	encodedToken, err := token.Encode()
	if err != nil {
		return response.InternalError(err)
	}

	return response.SyncResponse(true, types.RemoteClusterTokenPostResponse{Token: encodedToken})
}

func tokenGet(s microState.State, r *http.Request) response.Response {
	var tokens []database.CoreRemoteClusterToken
	err := s.Database().Transaction(r.Context(), func(ctx context.Context, tx *sql.Tx) error {
		var err error
		tokens, err = database.GetCoreRemoteClusterTokens(ctx, tx)
		return err
	})

	if err != nil {
		return response.SmartError(err)
	}

	var responseTokens []types.RemoteClusterToken
	for _, token := range tokens {
		responseTokens = append(responseTokens, types.RemoteClusterToken{
			Expiry:      token.Expiry,
			ClusterName: token.ClusterName,
			CreateAt:    token.CreatedAt,
		})
	}

	return response.SyncResponse(true, responseTokens)
}

func tokenDelete(s microState.State, r *http.Request) response.Response {
	remoteClusterName, err := url.PathUnescape(mux.Vars(r)["remoteClusterName"])
	if err != nil {
		return response.SmartError(err)
	}

	if remoteClusterName == "" {
		return response.BadRequest(fmt.Errorf("cluster name is required"))
	}

	err = s.Database().Transaction(r.Context(), func(ctx context.Context, tx *sql.Tx) error {
		return database.DeleteCoreRemoteClusterToken(ctx, tx, remoteClusterName)
	})

	if err != nil {
		return response.SmartError(err)
	}

	return response.SyncResponse(true, nil)
}

// getClusterManagerAddresses returns the addresses of the cluster managers that are online.
func getClusterManagerAddresses(ctx context.Context, s microState.State) ([]string, error) {
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

func getGlobalAddress(s microState.State) (string, error) {
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

func getMemberConfigs(s microState.State) ([]database.ManagerMemberConfig, error) {
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

// getLeaderPrioritisedAddresses returns the addresses of Cluster Manager members (external if any is set) with the leader as the first address.
func getLeaderPrioritisedAddresses(ctx context.Context, s microState.State, memberConfigs []database.ManagerMemberConfig) ([]string, error) {
	hasExternalAddresses := checkExternalAddresses(memberConfigs)
	leaderName, onlineMembers, err := getLeaderAndOnlineMembers(ctx, s)
	if err != nil {
		return nil, err
	}

	clusterManagerAddresses := make([]string, 0, len(memberConfigs))
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

		clusterManagerAddresses = append(clusterManagerAddresses, address)
	}

	if leaderIndex != 0 {
		clusterManagerAddresses[0], clusterManagerAddresses[leaderIndex] = clusterManagerAddresses[leaderIndex], clusterManagerAddresses[0]
	}

	return append(clusterManagerAddresses, nonOnlineMembers...), nil
}

func checkExternalAddresses(memberConfigs []database.ManagerMemberConfig) bool {
	for _, memberConfig := range memberConfigs {
		if memberConfig.ExternalAddress != "" {
			return true
		}
	}

	return false
}

func getLeaderAndOnlineMembers(ctx context.Context, s microState.State) (leaderName string, onlineMembers map[string]bool, err error) {
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

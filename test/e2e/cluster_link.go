package main

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"testing"

	"github.com/canonical/lxd/shared/api"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/api/models/v1"
	"github.com/canonical/microcloud-cluster-manager/test/helpers"
)

// testClusterLinkBadRequests checks request validation without changing LXD grants.
func testClusterLinkBadRequests(env *helpers.Environment) (testName string, testFunc func(t *testing.T)) {
	return "cluster link creation rejects bad requests before creating anything", func(t *testing.T) {
		cert, err := env.ManagementAPICert().PublicKeyX509()
		helpers.LogTestOutcome(t, "Should read the management API certificate", err)
		headers, err := env.ManagementAPILoginHeaders(cert)
		helpers.LogTestOutcome(t, "Should log in to the management API", err)
		for _, clusterName := range []string{tunnelSourceClusterName, tunnelTargetClusterName} {
			checkClusterTunnel(t, env, clusterName)
		}

		cases := []struct {
			name          string
			request       models.RemoteClusterLinkPost
			expectedError string
		}{
			{
				name:          "unsupported link type",
				request:       models.RemoteClusterLinkPost{Type: "bogus", TargetCluster: tunnelTargetClusterName},
				expectedError: "Unsupported cluster link type",
			},
			{
				name:          "missing link type",
				request:       models.RemoteClusterLinkPost{TargetCluster: tunnelTargetClusterName},
				expectedError: "Unsupported cluster link type",
			},
			{
				name:          "target equals source",
				request:       models.RemoteClusterLinkPost{Type: "bidirectional", TargetCluster: tunnelSourceClusterName},
				expectedError: "Source cluster and target cluster must differ",
			},
			{
				name:          "missing target",
				request:       models.RemoteClusterLinkPost{Type: "bidirectional"},
				expectedError: "Missing target cluster",
			},
		}
		for _, c := range cases {
			passed := t.Run(c.name, func(t *testing.T) {
				before, name, reverseName := prepareClusterLinkTest(t, env, headers)
				t.Cleanup(func() {
					cleanupClusterLinkTest(t, env, before, name, reverseName, headers)
				})
				request := c.request
				request.Name = name
				path := api.NewURL().Scheme("https").Host(env.ManagementAPIHostPort()).Path("1.0", "remote-cluster", tunnelSourceClusterName, "cluster-links")
				status, requestErr := helpers.QueryManagementAPI(env, http.MethodPost, path, request, nil, headers)
				after := readClusterLinkSnapshots(t, env, headers)

				var errs []error
				if status != http.StatusBadRequest || requestErr == nil || !strings.Contains(requestErr.Error(), c.expectedError) {
					errs = append(errs, fmt.Errorf("expected status 400 with %q, got status %d: %v", c.expectedError, status, requestErr))
				}
				for clusterName, state := range before {
					if !reflect.DeepEqual(state, after[clusterName]) {
						errs = append(errs, fmt.Errorf("rejected creation changed links or identities on cluster %q", clusterName))
					}
				}
				helpers.LogTestOutcome(t, fmt.Sprintf("Case %q should reject a bad request without changing either cluster's links or identities", c.name), errors.Join(errs...))
			})
			if !passed {
				helpers.LogTestOutcome(t, "Should complete the case and cleanup before running another case", fmt.Errorf("cluster link case %q failed", c.name))
			}
		}
	}
}

// testClusterLinkCreateePrechecks checks missing and minimum LXD permissions.
// Each case snapshots both clusters, restricts grants, attempts creation,
// restores grants for observation, checks the result, and cleans up test resources.
func testClusterLinkCreateePrechecks(env *helpers.Environment) (testName string, testFunc func(t *testing.T)) {
	return "cluster link creation checks LXD permissions before creating anything", func(t *testing.T) {
		cert, err := env.ManagementAPICert().PublicKeyX509()
		helpers.LogTestOutcome(t, "Should read the management API certificate", err)
		headers, err := env.ManagementAPILoginHeaders(cert)
		helpers.LogTestOutcome(t, "Should log in to the management API", err)
		for _, clusterName := range []string{tunnelSourceClusterName, tunnelTargetClusterName} {
			checkClusterTunnel(t, env, clusterName)
		}

		full := []string{"can_create_cluster_links", "can_create_identities"}
		linksOnly := []string{"can_create_cluster_links"}
		identitiesOnly := []string{"can_create_identities"}
		bidirectional := models.RemoteClusterLinkPost{Type: "bidirectional", TargetCluster: tunnelTargetClusterName}
		unidirectional := models.RemoteClusterLinkPost{Type: "unidirectional", TargetCluster: tunnelTargetClusterName}
		public := models.RemoteClusterLinkPost{Type: "public", TargetCluster: tunnelTargetClusterName}

		cases := []struct {
			name              string
			request           models.RemoteClusterLinkPost
			sourcePermissions []string
			targetPermissions []string
			deniedCluster     string
			deniedEntitlement string
		}{
			{
				name: "bidirectional source missing links", request: bidirectional,
				sourcePermissions: identitiesOnly, targetPermissions: full,
				deniedCluster: tunnelSourceClusterName, deniedEntitlement: "can_create_cluster_links",
			},
			{
				name: "bidirectional source missing identities", request: bidirectional,
				sourcePermissions: linksOnly, targetPermissions: full,
				deniedCluster: tunnelSourceClusterName, deniedEntitlement: "can_create_identities",
			},
			{
				name: "bidirectional target missing links", request: bidirectional,
				sourcePermissions: full, targetPermissions: identitiesOnly,
				deniedCluster: tunnelTargetClusterName, deniedEntitlement: "can_create_cluster_links",
			},
			{
				name: "bidirectional target missing identities", request: bidirectional,
				sourcePermissions: full, targetPermissions: linksOnly,
				deniedCluster: tunnelTargetClusterName, deniedEntitlement: "can_create_identities",
			},
			{
				name: "unidirectional source missing links", request: unidirectional,
				targetPermissions: full,
				deniedCluster:     tunnelSourceClusterName, deniedEntitlement: "can_create_cluster_links",
			},
			{
				name: "unidirectional target missing links", request: unidirectional,
				sourcePermissions: linksOnly, targetPermissions: identitiesOnly,
				deniedCluster: tunnelTargetClusterName, deniedEntitlement: "can_create_cluster_links",
			},
			{
				name: "unidirectional target missing identities", request: unidirectional,
				sourcePermissions: linksOnly, targetPermissions: linksOnly,
				deniedCluster: tunnelTargetClusterName, deniedEntitlement: "can_create_identities",
			},
			{
				name: "public source missing links", request: public,
				deniedCluster: tunnelSourceClusterName, deniedEntitlement: "can_create_cluster_links",
			},
			{
				name: "bidirectional minimum permissions", request: bidirectional,
				sourcePermissions: full, targetPermissions: full,
			},
			{
				name: "unidirectional minimum permissions", request: unidirectional,
				sourcePermissions: linksOnly, targetPermissions: full,
			},
			{
				name: "public minimum permissions", request: public,
				sourcePermissions: linksOnly,
			},
		}
		for _, c := range cases {
			passed := t.Run(c.name, func(t *testing.T) {
				before, name, reverseName := prepareClusterLinkTest(t, env, headers)
				t.Cleanup(func() {
					cleanupClusterLinkTest(t, env, before, name, reverseName, headers)
				})

				// Replace LXD IdP grants with exactly the listed permissions
				// (nil means none), while keeping the MCM admin session.
				var restore []func() error
				for _, cluster := range []struct {
					name        string
					permissions []string
				}{
					{name: tunnelSourceClusterName, permissions: c.sourcePermissions},
					{name: tunnelTargetClusterName, permissions: c.targetPermissions},
				} {
					restoreMapping, cleanup, err := restrictClusterLinkPermissions(cluster.name, cluster.permissions)
					if cleanup != nil {
						t.Cleanup(func() {
							helpers.LogTestOutcome(t, fmt.Sprintf("Should restore IdP mappings and delete the temporary group on cluster %q", cluster.name), cleanup())
						})
					}
					helpers.LogTestOutcome(t, fmt.Sprintf("Should restrict IdP permissions on cluster %q", cluster.name), err)
					restore = append(restore, restoreMapping)
				}

				request := c.request
				request.Name = name
				// Attempt creation once through the real reverse tunnels.
				path := api.NewURL().Scheme("https").Host(env.ManagementAPIHostPort()).Path("1.0", "remote-cluster", tunnelSourceClusterName, "cluster-links")
				status, requestErr := helpers.QueryManagementAPI(env, http.MethodPost, path, request, nil, headers)

				// Observe with the original grants, without deleting evidence of a failed preflight.
				for _, restoreMapping := range restore {
					err := restoreMapping()
					helpers.LogTestOutcome(t, "Should restore original IdP mappings before observing results", err)
				}

				after := readClusterLinkSnapshots(t, env, headers)
				if c.deniedEntitlement != "" {
					expectedError := fmt.Sprintf("Insufficient permissions on cluster %q: missing %q entitlement", c.deniedCluster, c.deniedEntitlement)
					var errs []error
					if status != http.StatusForbidden || requestErr == nil || !strings.Contains(requestErr.Error(), expectedError) {
						errs = append(errs, fmt.Errorf("expected status 403 with %q, got status %d: %v", expectedError, status, requestErr))
					}
					for clusterName, state := range before {
						if !reflect.DeepEqual(state, after[clusterName]) {
							errs = append(errs, fmt.Errorf("rejected creation changed links or identities on cluster %q", clusterName))
						}
					}
					helpers.LogTestOutcome(t, fmt.Sprintf("Case %q should reject creation without changing either cluster's links or identities", c.name), errors.Join(errs...))
					return
				}

				var outcomeErr error
				if requestErr != nil || status < http.StatusOK || status >= http.StatusMultipleChoices {
					outcomeErr = fmt.Errorf("expected successful creation, got status %d: %v", status, requestErr)
				}
				helpers.LogTestOutcome(t, "Should create the cluster link with the required permissions", outcomeErr)
				assertClusterLinkActive(t, after[tunnelSourceClusterName], name, request.Type)
				if request.Type == "bidirectional" {
					assertClusterLinkActive(t, after[tunnelTargetClusterName], reverseName, request.Type)
				} else {
					var outcomeErr error
					if !reflect.DeepEqual(before[tunnelTargetClusterName].Links, after[tunnelTargetClusterName].Links) {
						outcomeErr = fmt.Errorf("one-way creation changed cluster links on the target")
					}
					helpers.LogTestOutcome(t, "Should leave target links unchanged for one-way creation", outcomeErr)
				}
			})
			if !passed {
				helpers.LogTestOutcome(t, "Should complete the case and cleanup before running another case", fmt.Errorf("cluster link case %q failed", c.name))
			}
		}
	}
}

// prepareClusterLinkTest snapshots both clusters and reserves unused test link names.
// The caller must register resource cleanup before restricting permissions.
func prepareClusterLinkTest(t *testing.T, env *helpers.Environment, headers func(*http.Request) error) (before map[string]clusterLinkSnapshot, name string, reverseName string) {
	before = readClusterLinkSnapshots(t, env, headers)
	name = helpers.GetRandomName("mcm-preflight")
	reverseName = tunnelSourceClusterName + "-link"
	_, exists := before[tunnelSourceClusterName].Links[name]
	if exists {
		helpers.LogTestOutcome(t, "Should use an unused source link name", fmt.Errorf("test link %q already exists", name))
	}
	_, exists = before[tunnelTargetClusterName].Links[reverseName]
	if exists {
		helpers.LogTestOutcome(t, "Should use an unused reciprocal link name", fmt.Errorf("reciprocal link %q already exists", reverseName))
	}
	for _, identity := range before[tunnelTargetClusterName].Identities {
		if identity.Name == reverseName {
			helpers.LogTestOutcome(t, "Should use an unused reciprocal identity name", fmt.Errorf("target already has an identity named %q", reverseName))
		}
	}
	return before, name, reverseName
}

// readClusterLinkSnapshots captures both clusters with their original LXD grants.
func readClusterLinkSnapshots(t *testing.T, env *helpers.Environment, headers func(*http.Request) error) map[string]clusterLinkSnapshot {
	snapshots := make(map[string]clusterLinkSnapshot, 2)
	for _, clusterName := range []string{tunnelSourceClusterName, tunnelTargetClusterName} {
		snapshot, err := readClusterLinkSnapshot(env, clusterName, headers)
		helpers.LogTestOutcome(t, fmt.Sprintf("Should read links and identities on cluster %q", clusterName), err)
		snapshots[clusterName] = snapshot
	}
	return snapshots
}

// clusterLinkSnapshot records links and identity membership for side-effect checks.
// Identity values retain only the name and type; map keys identify method and identifier.
type clusterLinkSnapshot struct {
	Links      map[string]api.ClusterLink
	Identities map[string]api.Identity
}

// readClusterLinkSnapshot reads through MCM with the original LXD grants. It omits
// identity certificate and expiry metadata that may change independently of creation,
// and rejects direct OIDC group grants that would bypass the test-controlled IdP mappings.
func readClusterLinkSnapshot(env *helpers.Environment, clusterName string, headers func(*http.Request) error) (clusterLinkSnapshot, error) {
	var links []api.ClusterLink
	path := api.NewURL().Scheme("https").Host(env.ManagementAPIHostPort()).Path("1.0", "remote-cluster", clusterName, "cluster-links")
	status, err := helpers.QueryManagementAPI(env, http.MethodGet, path, nil, &links, headers)
	if err != nil || status != http.StatusOK || links == nil {
		return clusterLinkSnapshot{}, fmt.Errorf("list links on cluster %q (status %d): %v", clusterName, status, err)
	}
	var identities []api.Identity
	path = api.NewURL().Scheme("https").Host(env.ManagementAPIHostPort()).Path("1.0", "remote-cluster", clusterName, "auth", "identities")
	status, err = helpers.QueryManagementAPI(env, http.MethodGet, path, nil, &identities, headers)
	if err != nil || status != http.StatusOK || identities == nil {
		return clusterLinkSnapshot{}, fmt.Errorf("list identities on cluster %q (status %d): %v", clusterName, status, err)
	}

	snapshot := clusterLinkSnapshot{
		Links:      make(map[string]api.ClusterLink, len(links)),
		Identities: make(map[string]api.Identity, len(identities)),
	}
	for _, link := range links {
		snapshot.Links[link.Name] = link
	}
	// Ignore certificate and expiry metadata that can change without creating an identity.
	for _, identity := range identities {
		if identity.AuthenticationMethod == api.AuthenticationMethodOIDC && len(identity.Groups) > 0 {
			return clusterLinkSnapshot{}, fmt.Errorf("OIDC identity on cluster %q has direct group grants outside the test-controlled IdP mappings", clusterName)
		}
		key := identity.AuthenticationMethod + "/" + identity.Identifier
		snapshot.Identities[key] = api.Identity{Name: identity.Name, Type: identity.Type}
	}
	return snapshot, nil
}

// restrictClusterLinkPermissions replaces every IdP mapping's LXD groups with one
// temporary group granting only the requested server entitlements. The returned function
// restores the original mappings so result snapshots can see all links and identities.
// Once a temporary group exists, cleanup is returned even if restriction fails.
// The caller must register it before handling err; it restores mappings, then deletes
// the temporary group. Both callbacks are nil if no temporary group was created.
func restrictClusterLinkPermissions(clusterName string, entitlements []string) (restore func() error, cleanup func() error, err error) {
	var mappings []api.IdentityProviderGroup
	err = helpers.QueryClusterLXD(clusterName, http.MethodGet, "/1.0/auth/identity-provider-groups?recursion=1", nil, &mappings)
	if err != nil {
		return nil, nil, fmt.Errorf("read IdP mappings on cluster %q: %w", clusterName, err)
	}
	if len(mappings) == 0 {
		return nil, nil, fmt.Errorf("cluster %q has no IdP mappings to restrict", clusterName)
	}
	groupName := helpers.GetRandomName("mcm-preflight")
	group := api.AuthGroupsPost{Name: groupName}
	for _, entitlement := range entitlements {
		group.Permissions = append(group.Permissions, api.Permission{
			EntityType: "server", EntityReference: "/1.0", Entitlement: entitlement,
		})
	}
	err = helpers.QueryClusterLXD(clusterName, http.MethodPost, "/1.0/auth/groups", group, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("create temporary permission group on cluster %q: %w", clusterName, err)
	}

	mappingsRestored := false
	restore = func() error {
		if mappingsRestored {
			return nil
		}
		var errs []error
		for _, mapping := range mappings {
			path := "/1.0/auth/identity-provider-groups/" + url.PathEscape(mapping.Name)
			err := helpers.QueryClusterLXD(clusterName, http.MethodPut, path, mapping.Writable(), nil)
			if err != nil {
				errs = append(errs, fmt.Errorf("restore IdP mapping %q on cluster %q: %w", mapping.Name, clusterName, err))
			}
		}
		err := errors.Join(errs...)
		mappingsRestored = err == nil
		return err
	}
	cleanup = func() error {
		err := restore()
		if err != nil {
			return err
		}
		err = helpers.QueryClusterLXD(clusterName, http.MethodDelete, "/1.0/auth/groups/"+url.PathEscape(groupName), nil, nil)
		if err != nil {
			return fmt.Errorf("delete temporary group %q on cluster %q: %w", groupName, clusterName, err)
		}
		return nil
	}
	for _, mapping := range mappings {
		path := "/1.0/auth/identity-provider-groups/" + url.PathEscape(mapping.Name)
		err := helpers.QueryClusterLXD(clusterName, http.MethodPut, path, api.IdentityProviderGroupPut{Groups: []string{groupName}}, nil)
		if err != nil {
			return restore, cleanup, fmt.Errorf("restrict IdP mapping %q on cluster %q: %w", mapping.Name, clusterName, err)
		}
	}
	return restore, cleanup, nil
}

// assertClusterLinkActive checks the link type, usable addresses, and absence of
// pending handshake metadata.
func assertClusterLinkActive(t *testing.T, snapshot clusterLinkSnapshot, name string, linkType string) {
	link, exists := snapshot.Links[name]
	active := link.Config["volatile.addresses"] != "" &&
		link.Config["volatile.pending_address"] == "" &&
		link.Config["volatile.pending_certificate"] == ""
	var err error
	if !exists || link.Type != linkType || !active {
		err = fmt.Errorf("expected active %s link %q", linkType, name)
	}
	helpers.LogTestOutcome(t, fmt.Sprintf("Should have an active %s link %q", linkType, name), err)
}

// cleanupClusterLinkTest removes only test-named links and TLS identities absent from
// the initial snapshot, then verifies both clusters match their original state.
// Register it before permission cleanup so LIFO cleanup restores grants first.
func cleanupClusterLinkTest(t *testing.T, env *helpers.Environment, before map[string]clusterLinkSnapshot, name string, reverseName string, headers func(*http.Request) error) {
	var errs []error
	for _, clusterName := range []string{tunnelSourceClusterName, tunnelTargetClusterName} {
		linkName := name
		if clusterName == tunnelTargetClusterName {
			linkName = reverseName
		}
		snapshot, err := readClusterLinkSnapshot(env, clusterName, headers)
		if err != nil {
			errs = append(errs, fmt.Errorf("read cleanup state on cluster %q: %w", clusterName, err))
			continue
		}
		_, existed := before[clusterName].Links[linkName]
		_, exists := snapshot.Links[linkName]
		if exists && !existed {
			path := api.NewURL().Scheme("https").Host(env.ManagementAPIHostPort()).Path("1.0", "remote-cluster", clusterName, "cluster-links", linkName)
			status, err := helpers.QueryManagementAPI(env, http.MethodDelete, path, nil, nil, headers)
			if err != nil || status < http.StatusOK || status >= http.StatusMultipleChoices {
				errs = append(errs, fmt.Errorf("delete test link on cluster %q (status %d): %v", clusterName, status, err))
			}
		}

		snapshot, err = readClusterLinkSnapshot(env, clusterName, headers)
		if err != nil {
			errs = append(errs, fmt.Errorf("read identity cleanup state on cluster %q: %w", clusterName, err))
			continue
		}
		for key, identity := range snapshot.Identities {
			_, existed := before[clusterName].Identities[key]
			if existed {
				continue
			}
			if !strings.HasPrefix(key, api.AuthenticationMethodTLS+"/") || identity.Name != linkName {
				continue
			}
			path := "/1.0/auth/identities/tls/" + url.PathEscape(strings.TrimPrefix(key, api.AuthenticationMethodTLS+"/"))
			err := helpers.QueryClusterLXD(clusterName, http.MethodDelete, path, nil, nil)
			if err != nil {
				errs = append(errs, fmt.Errorf("delete test identity on cluster %q: %w", clusterName, err))
			}
		}
		final, err := readClusterLinkSnapshot(env, clusterName, headers)
		if err != nil {
			errs = append(errs, fmt.Errorf("read final cleanup state on cluster %q: %w", clusterName, err))
		} else if !reflect.DeepEqual(before[clusterName], final) {
			errs = append(errs, fmt.Errorf("cluster %q did not return to its original links and identities", clusterName))
		}
	}
	helpers.LogTestOutcome(t, "Should restore both clusters' original links and identities", errors.Join(errs...))
}

// testClusterLinkUnreachable keeps peers without tunnels to cover unreachable-cluster preflight.
func testClusterLinkUnreachable(env *helpers.Environment) (testName string, testFunc func(t *testing.T)) {
	return "cluster link creation rejects unreachable clusters before creating anything", func(t *testing.T) {
		sourceName := helpers.GetRandomName("cluster_link_e2e_src")
		targetName := helpers.GetRandomName("cluster_link_e2e_tgt")

		defer func() {
			env.RemoveRemoteCluster(sourceName)
			env.RemoveRemoteCluster(targetName)
			env.RemoveRemoteClusterToken(sourceName)
			env.RemoveRemoteClusterToken(targetName)
		}()

		for _, clusterName := range []string{sourceName, targetName} {
			_, err := helpers.RegisterRemoteCluster(env, clusterName)
			helpers.LogTestOutcome(t, fmt.Sprintf("Should register unreachable cluster %q", clusterName), err)
		}

		// The internal cluster connector endpoints expect the headers the management API
		// sets when forwarding a user request.
		internalHeaders := func(req *http.Request) error {
			req.Header.Set("Authorization", "Bearer e2e-test-token")
			req.Header.Set("X-User-Secret", "e2e-test-secret")
			return nil
		}

		path := api.NewURL().Scheme("https").Host(env.ClusterConnectorHostPort()).Path("1.0", "remote-cluster", sourceName, "cluster-links")
		input := models.RemoteClusterLinkPost{TargetCluster: targetName, Type: "bidirectional"}
		status, err := helpers.QueryClusterConnectorInternal(env, http.MethodPost, path, input, nil, internalHeaders)
		var outcomeErr error
		if status != http.StatusBadGateway || err == nil || !strings.Contains(err.Error(), "unreachable") {
			outcomeErr = fmt.Errorf("expected status 502 with an unreachable-cluster error, got status %d: %v", status, err)
		}
		helpers.LogTestOutcome(t, "Should reject an unreachable cluster before creating a pending link", outcomeErr)
	}
}

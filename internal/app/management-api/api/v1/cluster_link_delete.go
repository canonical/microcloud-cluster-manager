package v1

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/canonical/lxd/shared"
	"github.com/canonical/lxd/shared/api"
	"github.com/canonical/microcloud-cluster-manager/internal/app/management-api/core/cluster_link"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/database/store"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/types"
)

// deleteClusterLinkWithTargetCleanup removes the target-side counterpart of a cluster link, then the link itself.
func deleteClusterLinkWithTargetCleanup(rc types.RouteConfig, r *http.Request, clusters []store.RemoteClusterWithDetail, sourceName string, linkName string) error {
	source := cluster_link.FindCluster(clusters, sourceName)
	if source == nil {
		return api.StatusErrorf(http.StatusNotFound, "Remote cluster %q not found", sourceName)
	}

	// Step 1: load the source link and fail early if the source deletion would fail anyway.
	var link api.ClusterLink
	err := requestClusterMember(rc, r, http.MethodGet, *source, "/cluster-links/"+url.PathEscape(linkName), &link, cluster_link.MaxResponseBytes)
	if err != nil {
		return err
	}
	if len(link.UsedBy) > 0 {
		return api.StatusErrorf(http.StatusConflict, "Cluster link %q on cluster %q is in use", linkName, source.Name)
	}

	if !cluster_link.ClusterLinkTypeSupportsTargetCleanup(link.Type) {
		return api.StatusErrorf(http.StatusBadRequest, "Target cleanup is supported only for %s and %s cluster links", api.ClusterLinkTypeBidirectional, api.ClusterLinkTypeUnidirectional)
	}

	// Step 2: resolve the target cluster from the UUID recorded on the link.
	target := findClusterByUUID(clusters, link.Config["volatile.uuid"])
	if target == nil || target.Name == source.Name {
		return api.StatusErrorf(http.StatusConflict, "Could not identify the target cluster for link %q on cluster %q. Retry later or delete the source link separately", linkName, source.Name)
	}

	// Step 3: remove the target-side counterpart.
	if strings.EqualFold(link.Type, api.ClusterLinkTypeBidirectional) {
		err = deleteReverseClusterLink(rc, r, *source, *target)
	} else {
		err = deleteClusterLinkIdentity(rc, r, *source, *target)
	}
	if err != nil {
		return fmt.Errorf("Target cleanup on cluster %q failed before source deletion: %w", target.Name, err)
	}

	// Step 4: delete the source link.
	err = requestClusterMember(rc, r, http.MethodDelete, *source, clusterLinksConnectorPath(source.Name, linkName), nil, cluster_link.MaxResponseBytes)
	if err != nil {
		return fmt.Errorf("Target cleanup on cluster %q completed, but deletion of source link %q on cluster %q failed: %w", target.Name, linkName, source.Name, err)
	}

	return nil
}

// findClusterByUUID returns the managed cluster with the given UUID, or nil.
func findClusterByUUID(clusters []store.RemoteClusterWithDetail, uuid string) *store.RemoteClusterWithDetail {
	if uuid == "" {
		return nil
	}

	for i := range clusters {
		if clusters[i].ClusterUUID == uuid {
			return &clusters[i]
		}
	}

	return nil
}

// deleteReverseClusterLink deletes the bidirectional link on the target that points back at the source, if any.
func deleteReverseClusterLink(rc types.RouteConfig, r *http.Request, source store.RemoteClusterWithDetail, target store.RemoteClusterWithDetail) error {
	var links []api.ClusterLink
	err := requestClusterMember(rc, r, http.MethodGet, target, clusterLinksConnectorPath(target.Name, ""), &links, cluster_link.MaxResponseBytes)
	if err != nil {
		return err
	}

	reverseLink, unresolved := cluster_link.FindReverseLinkByUUID(links, source.ClusterUUID)
	if unresolved {
		return api.StatusErrorf(http.StatusConflict, "Could not verify whether a reverse link points to cluster %q. Retry later or delete the source link separately", source.Name)
	}
	if reverseLink == nil {
		return nil
	}

	err = requestClusterMember(rc, r, http.MethodDelete, target, clusterLinksConnectorPath(target.Name, reverseLink.Name), nil, cluster_link.MaxResponseBytes)
	if api.StatusErrorCheck(err, http.StatusNotFound) {
		return nil
	}

	return err
}

// deleteClusterLinkIdentity deletes the TLS identity of the source cluster on the target, if any.
func deleteClusterLinkIdentity(rc types.RouteConfig, r *http.Request, source store.RemoteClusterWithDetail, target store.RemoteClusterWithDetail) error {
	var sourceInfo api.Server
	err := requestClusterMember(rc, r, http.MethodGet, source, "/info", &sourceInfo, cluster_link.MaxResponseBytes)
	if err != nil {
		return err
	}

	fingerprint := sourceInfo.Environment.CertificateFingerprint
	if fingerprint == "" && sourceInfo.Environment.Certificate != "" {
		fingerprint, err = shared.CertFingerprintStr(sourceInfo.Environment.Certificate)
		if err != nil {
			return api.StatusErrorf(http.StatusBadGateway, "Cluster %q reported an invalid certificate: %v", source.Name, err)
		}
	}
	if fingerprint == "" {
		return api.StatusErrorf(http.StatusBadGateway, "Cluster %q did not report a certificate or fingerprint", source.Name)
	}

	var identities []api.Identity
	err = requestClusterMember(rc, r, http.MethodGet, target, "/auth/identities", &identities, cluster_link.MaxResponseBytes)
	if err != nil {
		return err
	}

	identity := cluster_link.FindIdentity(identities, fingerprint)
	if identity == nil {
		return nil
	}

	err = requestClusterMember(rc, r, http.MethodDelete, target, "/auth/identities/tls/"+url.PathEscape(identity.Identifier), nil, cluster_link.MaxResponseBytes)
	if api.StatusErrorCheck(err, http.StatusNotFound) {
		return nil
	}

	return err
}

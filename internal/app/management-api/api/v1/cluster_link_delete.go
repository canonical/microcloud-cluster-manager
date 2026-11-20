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
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/logger"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/types"
)

// deleteClusterLinkOnBothClusters deletes the named cluster link on its source cluster, after deleting what the link
// relies on on its target cluster: the reverse link for bidirectional links, or the link identity for unidirectional
// links. The link type and the target cluster are resolved from the link on the source cluster and the managed clusters.
func deleteClusterLinkOnBothClusters(rc types.RouteConfig, r *http.Request, clusters []store.RemoteClusterWithDetail, sourceClusterName string, linkName string) error {
	sourceCluster := cluster_link.FindCluster(clusters, sourceClusterName)
	if sourceCluster == nil {
		return api.StatusErrorf(http.StatusNotFound, "remote cluster %q not found", sourceClusterName)
	}

	var sourceLinks []api.ClusterLink
	err := requestClusterMember(rc, r, http.MethodGet, *sourceCluster, clusterLinksConnectorPath, &sourceLinks, cluster_link.MaxResponseBytes)
	if err != nil {
		return err
	}

	sourceLink := cluster_link.FindLinkByName(sourceLinks, linkName)
	if sourceLink == nil {
		return api.StatusErrorf(http.StatusNotFound, "cluster link %q not found on cluster %q", linkName, sourceClusterName)
	}

	linkType := strings.ToLower(sourceLink.Type)
	if linkType != api.ClusterLinkTypeBidirectional && linkType != api.ClusterLinkTypeUnidirectional {
		return api.StatusErrorf(http.StatusBadRequest, "deleting a cluster link on both clusters is only supported for %s and %s cluster links", api.ClusterLinkTypeBidirectional, api.ClusterLinkTypeUnidirectional)
	}

	targetCluster := cluster_link.NewTargetIndex(clusters).Resolve(*sourceLink, sourceClusterName)
	if targetCluster == nil {
		return api.StatusErrorf(http.StatusBadRequest, "cluster link %q on cluster %q does not point at a registered cluster", linkName, sourceClusterName)
	}

	targetClusterName := targetCluster.Name

	// Clean up the target cluster before deleting the source link. A failure on the target cluster then leaves the
	// source link in place, so the deletion can be retried. An already missing reverse link or link identity counts as
	// deleted, so a retry also succeeds when only the source link deletion failed.
	if linkType == api.ClusterLinkTypeBidirectional {
		err = deleteReverseClusterLink(rc, r, *sourceCluster, *targetCluster)
	} else {
		err = deleteClusterLinkIdentity(rc, r, *sourceCluster, *targetCluster)
	}

	if err != nil {
		logger.Log.Errorw("Failed to clean up the target cluster of a cluster link", "source_cluster", sourceClusterName, "target_cluster", targetClusterName, "cluster_link", linkName, "type", linkType, "error", err)
		return fmt.Errorf("failed to clean up target cluster %q, cluster link %q on cluster %q was not deleted: %w", targetClusterName, linkName, sourceClusterName, err)
	}

	return requestClusterMember(rc, r, http.MethodDelete, *sourceCluster, cluster_link.LinkPath(linkName), nil, cluster_link.MaxResponseBytes)
}

// deleteReverseClusterLink deletes the bidirectional cluster link on the target cluster that points back at the
// source cluster. A missing reverse link is treated as already deleted.
func deleteReverseClusterLink(rc types.RouteConfig, r *http.Request, sourceCluster store.RemoteClusterWithDetail, targetCluster store.RemoteClusterWithDetail) error {
	var targetLinks []api.ClusterLink
	err := requestClusterMember(rc, r, http.MethodGet, targetCluster, clusterLinksConnectorPath, &targetLinks, cluster_link.MaxResponseBytes)
	if err != nil {
		return err
	}

	reverseLink := cluster_link.FindReverseLink(targetLinks, sourceCluster.ClusterUUID, sourceCluster.LXDURL)
	if reverseLink == nil {
		logger.Log.Infow("No reverse cluster link found on target cluster, assuming it was already deleted", "source_cluster", sourceCluster.Name, "target_cluster", targetCluster.Name)
		return nil
	}

	err = requestClusterMember(rc, r, http.MethodDelete, targetCluster, cluster_link.LinkPath(reverseLink.Name), nil, cluster_link.MaxResponseBytes)
	if api.StatusErrorCheck(err, http.StatusNotFound) {
		return nil
	}

	return err
}

// deleteClusterLinkIdentity deletes the cluster link identity on the target cluster that holds the certificate of
// the source cluster. A missing identity is treated as already deleted.
func deleteClusterLinkIdentity(rc types.RouteConfig, r *http.Request, sourceCluster store.RemoteClusterWithDetail, targetCluster store.RemoteClusterWithDetail) error {
	var sourceServer api.Server
	err := requestClusterMember(rc, r, http.MethodGet, sourceCluster, "/info", &sourceServer, cluster_link.MaxResponseBytes)
	if err != nil {
		return err
	}

	fingerprint := sourceServer.Environment.CertificateFingerprint
	if fingerprint == "" && sourceServer.Environment.Certificate != "" {
		fingerprint, err = shared.CertFingerprintStr(sourceServer.Environment.Certificate)
		if err != nil {
			return api.StatusErrorf(http.StatusBadGateway, "cluster %q reported an invalid certificate: %v", sourceCluster.Name, err)
		}
	}

	if fingerprint == "" {
		return api.StatusErrorf(http.StatusBadGateway, "cluster %q did not report a certificate", sourceCluster.Name)
	}

	var identities []api.Identity
	err = requestClusterMember(rc, r, http.MethodGet, targetCluster, "/auth/identities", &identities, cluster_link.MaxResponseBytes)
	if err != nil {
		return err
	}

	identity := cluster_link.FindIdentity(identities, fingerprint)
	if identity == nil {
		logger.Log.Infow("No cluster link identity found on target cluster, assuming it was already deleted", "source_cluster", sourceCluster.Name, "target_cluster", targetCluster.Name)
		return nil
	}

	err = requestClusterMember(rc, r, http.MethodDelete, targetCluster, "/auth/identities/tls/"+url.PathEscape(identity.Identifier), nil, cluster_link.MaxResponseBytes)
	if api.StatusErrorCheck(err, http.StatusNotFound) {
		return nil
	}

	return err
}

package cluster_link

import (
	"net"
	"net/url"
	"strconv"
	"strings"

	"github.com/canonical/lxd/shared"
	"github.com/canonical/lxd/shared/api"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/api/models/v1"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/database/store"
)

// TargetIndex finds the managed cluster a cluster link points at. Build it once per request, so that resolving the
// target of a link is a lookup instead of a scan of all managed clusters.
type TargetIndex struct {
	byUUID    map[string]*store.RemoteClusterWithDetail
	byAddress map[string]*store.RemoteClusterWithDetail
}

// NewTargetIndex indexes managed clusters by UUID and LXD address for list responses.
func NewTargetIndex(clusters []store.RemoteClusterWithDetail) TargetIndex {
	index := TargetIndex{
		byUUID:    make(map[string]*store.RemoteClusterWithDetail, len(clusters)),
		byAddress: make(map[string]*store.RemoteClusterWithDetail, len(clusters)),
	}

	for i := range clusters {
		cluster := &clusters[i]
		if cluster.ClusterUUID != "" && index.byUUID[cluster.ClusterUUID] == nil {
			index.byUUID[cluster.ClusterUUID] = cluster
		}

		address := lxdURLAddress(cluster.LXDURL)
		if address != "" && index.byAddress[address] == nil {
			index.byAddress[address] = cluster
		}
	}

	return index
}

// resolveListedTarget uses the link address for public links that have no target UUID.
func (i TargetIndex) resolveListedTarget(link api.ClusterLink, sourceCluster string) *store.RemoteClusterWithDetail {
	if uuid := link.Config["volatile.uuid"]; uuid != "" || !strings.EqualFold(link.Type, api.ClusterLinkTypePublic) {
		target := i.byUUID[uuid]
		if target != nil && target.Name != sourceCluster {
			return target
		}

		return nil
	}

	addresses := shared.SplitNTrimSpace(link.Config["volatile.addresses"], ",", -1, true)
	if pendingAddress := strings.TrimSpace(link.Config["volatile.pending_address"]); len(addresses) == 0 && pendingAddress != "" {
		addresses = []string{pendingAddress}
	}

	for _, address := range addresses {
		target := i.byAddress[strings.ToLower(address)]
		if target != nil && target.Name != sourceCluster {
			return target
		}
	}

	return nil
}

// lxdURLAddress returns the normalized host and port of a registered LXD URL.
func lxdURLAddress(lxdURL string) string {
	parsedURL, err := url.Parse(strings.TrimSpace(lxdURL))
	if err != nil || parsedURL.Hostname() == "" {
		return ""
	}

	port := parsedURL.Port()
	if port == "" {
		port = strconv.Itoa(shared.HTTPSDefaultPort)
	}

	return strings.ToLower(net.JoinHostPort(parsedURL.Hostname(), port))
}

// ResolveLinks returns the links defined on the named source cluster with their target clusters.
func ResolveLinks(sourceCluster string, links []api.ClusterLink, targets TargetIndex) []models.ResolvedClusterLink {
	resolved := make([]models.ResolvedClusterLink, 0, len(links))
	for _, link := range links {
		resolvedLink := models.ResolvedClusterLink{
			Link:          link,
			SourceCluster: sourceCluster,
		}

		target := targets.resolveListedTarget(link, sourceCluster)
		if target != nil {
			resolvedLink.TargetCluster = &target.Name
		}

		resolved = append(resolved, resolvedLink)
	}

	return resolved
}

// FindCluster returns the managed cluster with the given name, or nil.
func FindCluster(clusters []store.RemoteClusterWithDetail, name string) *store.RemoteClusterWithDetail {
	for i := range clusters {
		if clusters[i].Name == name {
			return &clusters[i]
		}
	}

	return nil
}

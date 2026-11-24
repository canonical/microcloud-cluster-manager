package cluster_link

import (
	"strings"

	"github.com/canonical/lxd/shared/api"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/api/models/v1"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/database/store"
)

// LinkAddresses returns the confirmed addresses of the link. A link without confirmed addresses, such as a pending
// public link, returns its pending address instead.
func LinkAddresses(link api.ClusterLink) []string {
	var addresses []string
	for _, address := range strings.Split(link.Config["volatile.addresses"], ",") {
		address = strings.TrimSpace(address)
		if address != "" {
			addresses = append(addresses, address)
		}
	}

	pendingAddress := strings.TrimSpace(link.Config["volatile.pending_address"])
	if len(addresses) == 0 && pendingAddress != "" {
		return []string{pendingAddress}
	}

	return addresses
}

// TargetIndex finds the managed cluster a cluster link points at. Build it once per request, so that resolving the
// target of a link is a lookup instead of a scan of all managed clusters.
type TargetIndex struct {
	byUUID    map[string]*store.RemoteClusterWithDetail
	byAddress map[string]*store.RemoteClusterWithDetail
}

// NewTargetIndex indexes the managed clusters by cluster UUID and by the address of their LXD URL. When several
// clusters share a UUID or an address, the first one wins.
func NewTargetIndex(clusters []store.RemoteClusterWithDetail) TargetIndex {
	index := TargetIndex{
		byUUID:    make(map[string]*store.RemoteClusterWithDetail, len(clusters)),
		byAddress: make(map[string]*store.RemoteClusterWithDetail, len(clusters)),
	}

	for i := range clusters {
		cluster := &clusters[i]
		_, found := index.byUUID[cluster.ClusterUUID]
		if cluster.ClusterUUID != "" && !found {
			index.byUUID[cluster.ClusterUUID] = cluster
		}

		address := lxdURLAddress(cluster.LXDURL)
		_, found = index.byAddress[address]
		if address != "" && !found {
			index.byAddress[address] = cluster
		}
	}

	return index
}

// Resolve returns the target cluster of a link defined on the named source cluster. A cluster whose UUID matches the
// link is preferred over a cluster whose LXD URL matches one of the link addresses. Returns nil if the link points at
// no managed cluster other than its source cluster.
func (i TargetIndex) Resolve(link api.ClusterLink, sourceCluster string) *store.RemoteClusterWithDetail {
	target := i.byUUID[link.Config["volatile.uuid"]]
	if target != nil && target.Name != sourceCluster {
		return target
	}

	for _, address := range LinkAddresses(link) {
		target = i.byAddress[strings.ToLower(address)]
		if target != nil && target.Name != sourceCluster {
			return target
		}
	}

	return nil
}

// ResolveLinks returns the links defined on the named source cluster with their target clusters.
func ResolveLinks(sourceCluster string, links []api.ClusterLink, targets TargetIndex) []models.ResolvedClusterLink {
	resolved := make([]models.ResolvedClusterLink, 0, len(links))
	for _, link := range links {
		resolvedLink := models.ResolvedClusterLink{
			Link:          link,
			SourceCluster: sourceCluster,
		}

		target := targets.Resolve(link, sourceCluster)
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

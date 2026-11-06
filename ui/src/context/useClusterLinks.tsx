import { useQuery } from "@tanstack/react-query";
import { queryKeys } from "util/queryKeys";
import type { Cluster, ClusterLink, ClusterLinkGroup } from "types/cluster";
import { fetchClusterLinkGroups } from "api/clusters";
import type { ClusterConnection } from "util/clusterLinks";
import {
  flattenClusterLinks,
  resolveClusterLinkTargets,
} from "util/clusterLinks";

interface UseClusterLinksResult {
  connections: ClusterConnection[];
  failedGroups: ClusterLinkGroup[];
  error: Error | null;
  isLoading: boolean;
}

// The management API groups the links by cluster and reports the clusters it
// failed to load next to the links it could load.
export const useClusterLinks = (clusters: Cluster[]): UseClusterLinksResult => {
  const {
    data: groups = [],
    error,
    isLoading,
  } = useQuery({
    queryKey: [queryKeys.clusters, queryKeys.links],
    queryFn: fetchClusterLinkGroups,
  });

  const linksByCluster: Record<string, ClusterLink[] | undefined> = {};
  for (const group of groups) {
    linksByCluster[group.cluster] = group.links;
  }

  const allLinks = flattenClusterLinks(clusters, linksByCluster);

  return {
    connections: resolveClusterLinkTargets(allLinks, clusters),
    failedGroups: groups.filter((group) => group.error),
    error,
    isLoading,
  };
};

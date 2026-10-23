import { useQueries } from "@tanstack/react-query";
import { queryKeys } from "util/queryKeys";
import type { Cluster, ClusterLink } from "types/cluster";
import { fetchClusterLinks } from "api/clusters";
import type { ClusterConnection } from "util/clusterLinks";
import { flattenClusterLinks } from "util/clusterLinks";

interface UseClusterLinksResult {
  connections: ClusterConnection[];
  failedClusters: Cluster[];
  isLoading: boolean;
}

// Fans out one query per tunnel-enabled cluster. The management API only
// exposes per-cluster link endpoints, so aggregation happens here; React Query
// caches and retries each cluster independently.
export const useClusterLinks = (
  sourceClusters: Cluster[],
  enrolledClusters: Cluster[] = sourceClusters,
): UseClusterLinksResult => {
  const tunnelClusters = sourceClusters.filter(
    (cluster) => cluster.tunnel_registered,
  );

  const linkQueries = useQueries({
    queries: tunnelClusters.map((cluster) => ({
      queryKey: [queryKeys.clusters, cluster.name, queryKeys.links],
      queryFn: async () => fetchClusterLinks(cluster.name),
    })),
  });

  const linksByCluster: Record<string, ClusterLink[] | undefined> = {};
  tunnelClusters.forEach((cluster, index) => {
    linksByCluster[cluster.name] = linkQueries[index]?.data;
  });

  const connections = flattenClusterLinks(
    sourceClusters,
    linksByCluster,
    enrolledClusters,
  );

  const failedClusters = tunnelClusters.filter(
    (_, index) => linkQueries[index]?.isError,
  );

  return {
    connections,
    failedClusters,
    isLoading: linkQueries.some((query) => query.isLoading),
  };
};

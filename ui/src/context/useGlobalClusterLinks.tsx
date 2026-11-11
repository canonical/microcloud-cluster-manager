import type { Cluster } from "types/cluster";
import type { ClusterConnection } from "util/clusterLinks";
import { getClusterConnections } from "util/clusterLinks";
import { useClustersWithLinks } from "context/useClusters";

interface UseGlobalClusterLinksResult {
  connections: ClusterConnection[];
  failedClusters: Cluster[];
  error: Error | null;
  isLoading: boolean;
}

// Clusters the API failed to load the links of are reported next to the links
// that could be loaded.
export const useGlobalClusterLinks = (): UseGlobalClusterLinksResult => {
  const {
    data: clusters = [] as Cluster[],
    error,
    isLoading,
  } = useClustersWithLinks();
  const connections = getClusterConnections(clusters);
  const failedClusters = clusters.filter(
    (cluster) => cluster.cluster_links_error,
  );

  return { connections, failedClusters, error, isLoading };
};

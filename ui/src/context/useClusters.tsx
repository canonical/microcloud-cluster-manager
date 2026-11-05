import type { UseQueryResult } from "@tanstack/react-query";
import { useQuery } from "@tanstack/react-query";
import { fetchCluster, fetchClusterLinks } from "api/clusters";
import type { Cluster, ClusterLink } from "types/cluster";
import { queryKeys } from "util/queryKeys";

export const useCluster = (clusterName: string): UseQueryResult<Cluster> => {
  return useQuery({
    queryKey: [queryKeys.clusters, clusterName],
    queryFn: async () => fetchCluster(clusterName),
    enabled: !!clusterName,
  });
};

export const useClusterLinks = (
  clusterName: string,
  enabled = true,
): UseQueryResult<ClusterLink[]> => {
  return useQuery({
    queryKey: [queryKeys.clusters, clusterName, queryKeys.links],
    queryFn: async () => fetchClusterLinks(clusterName),
    enabled: !!clusterName && enabled,
  });
};

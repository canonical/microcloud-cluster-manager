import type { UseQueryResult } from "@tanstack/react-query";
import { useQuery } from "@tanstack/react-query";
import {
  fetchAllClusterLinks,
  fetchCluster,
  fetchClusterLinks,
  fetchClusters,
} from "api/clusters";
import type {
  Cluster,
  ClusterLinkList,
  ResolvedClusterLink,
} from "types/cluster";
import { clusterLinkQueryKeys } from "util/clusterLinkQueries";
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
): UseQueryResult<ResolvedClusterLink[]> => {
  return useQuery({
    queryKey: clusterLinkQueryKeys.forCluster(clusterName),
    queryFn: async () => fetchClusterLinks(clusterName),
    enabled: !!clusterName && enabled,
  });
};

export const useAllClusterLinks = (): UseQueryResult<ClusterLinkList> => {
  return useQuery({
    queryKey: clusterLinkQueryKeys.all,
    queryFn: fetchAllClusterLinks,
  });
};

export const useClusters = (): UseQueryResult<Cluster[]> => {
  return useQuery({
    queryKey: [queryKeys.clusters],
    queryFn: fetchClusters,
  });
};

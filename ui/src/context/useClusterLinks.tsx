import type { UseQueryResult } from "@tanstack/react-query";
import { useQuery } from "@tanstack/react-query";
import {
  fetchAllClusterLinks,
  fetchClusterLinks,
  fetchClusterLinkState,
} from "api/clusters";
import type {
  ClusterLinkList,
  ClusterLinkState,
  ResolvedClusterLink,
} from "types/cluster";
import { allClusterLinksQueryKey, queryKeys } from "util/queryKeys";

export const useClusterLinks = (
  clusterName: string,
  enabled = true,
): UseQueryResult<ResolvedClusterLink[]> => {
  return useQuery({
    queryKey: [queryKeys.clusters, clusterName, queryKeys.links],
    queryFn: async () => fetchClusterLinks(clusterName),
    enabled: !!clusterName && enabled,
  });
};

export const useClusterLinkState = (
  clusterName: string,
  link: string,
): UseQueryResult<ClusterLinkState> => {
  return useQuery({
    queryKey: [
      queryKeys.clusters,
      clusterName,
      queryKeys.links,
      link,
      queryKeys.state,
    ],
    queryFn: async () => fetchClusterLinkState(clusterName, link),
  });
};

export const useAllClusterLinks = (): UseQueryResult<ClusterLinkList> => {
  return useQuery({
    queryKey: allClusterLinksQueryKey,
    queryFn: fetchAllClusterLinks,
  });
};

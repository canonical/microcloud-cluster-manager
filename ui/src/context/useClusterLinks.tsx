import type { UseQueryResult } from "@tanstack/react-query";
import { useQuery } from "@tanstack/react-query";
import {
  fetchClusterLink,
  fetchClusterLinks,
  fetchClusterLinkState,
} from "api/clusters";
import type { ClusterLink, ClusterLinkState } from "types/cluster";
import { queryKeys } from "util/queryKeys";

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

export const useClusterLink = (
  clusterName: string,
  link: string,
  enabled = true,
): UseQueryResult<ClusterLink> => {
  return useQuery({
    queryKey: [queryKeys.clusters, clusterName, queryKeys.links, link],
    queryFn: async () => fetchClusterLink(clusterName, link),
    enabled: !!clusterName && !!link && enabled,
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

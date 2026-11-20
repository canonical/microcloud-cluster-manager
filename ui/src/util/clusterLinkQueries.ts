import type { QueryClient, QueryKey } from "@tanstack/react-query";
import { queryKeys } from "util/queryKeys";

// Cluster-wide changes such as removal still invalidate link queries.
const clusterLinksKey = [
  queryKeys.clusters,
  { resource: queryKeys.links },
] as const;

export const clusterLinkQueryKeys = {
  all: [...clusterLinksKey, "all"] as const,
  forCluster: (clusterName: string): QueryKey => [
    ...clusterLinksKey,
    "cluster",
    clusterName,
  ],
};

export const invalidateClusterLinks = async (
  queryClient: QueryClient,
): Promise<void> => {
  await queryClient.invalidateQueries({ queryKey: clusterLinksKey });
};

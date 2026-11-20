import type { UseQueryResult } from "@tanstack/react-query";
import { useQuery } from "@tanstack/react-query";
import { fetchCluster } from "api/clusters";
import type { Cluster } from "types/cluster";
import { queryKeys } from "util/queryKeys";

export const useCluster = (clusterName: string): UseQueryResult<Cluster> => {
  return useQuery({
    queryKey: [queryKeys.clusters, clusterName],
    queryFn: async () => fetchCluster(clusterName),
    enabled: !!clusterName,
  });
};

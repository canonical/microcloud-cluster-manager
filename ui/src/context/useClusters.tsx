import type { UseQueryResult } from "@tanstack/react-query";
import { useQuery } from "@tanstack/react-query";
import { queryKeys } from "util/queryKeys";
import type { Cluster } from "types/cluster";
import { fetchClusters } from "api/clusters";

export const useClusters = (): UseQueryResult<Cluster[]> => {
  return useQuery({
    queryKey: [queryKeys.clusters],
    queryFn: fetchClusters,
  });
};

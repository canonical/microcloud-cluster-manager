import { useQuery, type UseQueryResult } from "@tanstack/react-query";
import { queryKeys } from "util/queryKeys";
import type { LxdIdentity } from "types/permissions";
import { fetchIdentities } from "api/auth-identities";

export const useClusterIdentities = (
  clusterName: string,
): UseQueryResult<LxdIdentity[]> => {
  return useQuery({
    queryKey: [queryKeys.clusters, clusterName, queryKeys.identities],
    queryFn: async () => fetchIdentities(clusterName),
    enabled: !!clusterName,
  });
};

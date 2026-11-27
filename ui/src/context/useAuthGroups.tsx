import { useQuery, type UseQueryResult } from "@tanstack/react-query";
import { queryKeys } from "util/queryKeys";
import type { LxdAuthGroup } from "types/permissions";
import { fetchGroups } from "api/auth-groups";

export const useAuthGroups = (
  clusterName: string,
  enabled = true,
): UseQueryResult<LxdAuthGroup[]> => {
  return useQuery({
    queryKey: [queryKeys.authGroups, clusterName],
    queryFn: async () => fetchGroups(clusterName),
    enabled,
  });
};

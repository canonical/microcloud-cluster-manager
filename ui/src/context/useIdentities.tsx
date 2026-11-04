import { useQuery, type UseQueryResult } from "@tanstack/react-query";
import { queryKeys } from "util/queryKeys";
import type { LxdIdentity } from "types/permissions";
import { fetchIdentities } from "api/auth-identities";

export const useIdentities = (
  cluster: string,
): UseQueryResult<LxdIdentity[]> => {
  return useQuery({
    queryKey: [queryKeys.identities, cluster],
    queryFn: async () => fetchIdentities(cluster),
    enabled: !!cluster,
  });
};

import type { UseQueryResult } from "@tanstack/react-query";
import { useQuery } from "@tanstack/react-query";
import { queryKeys } from "util/queryKeys";
import type { Server } from "types/server";
import { fetchServer } from "api/server";

export const useServer = (): UseQueryResult<Server> => {
  return useQuery({
    queryKey: [queryKeys.server],
    queryFn: fetchServer,
    retry: (count, error) => {
      const apiError = error as unknown as {
        response: { error_code?: number };
      };
      const statusCode = apiError?.response?.error_code ?? 0;
      const isServerError = statusCode >= 500 || statusCode === 0;
      return count < 3 && isServerError;
    },
  });
};

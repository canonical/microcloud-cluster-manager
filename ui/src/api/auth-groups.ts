import { ensureStableSorting, handleResponse } from "util/helpers";
import type { LxdApiResponse } from "types/apiResponse";
import type { LxdAuthGroup } from "types/permissions";
import { ROOT_PATH } from "util/rootPath";

export const groupEntitlements = ["can_delete", "can_edit"];

export const fetchGroups = async (
  remoteClusterName: string,
): Promise<LxdAuthGroup[]> => {
  const params = new URLSearchParams();
  params.set("recursion", "1");

  return fetch(
    `${ROOT_PATH}/1.0/remote-cluster/${encodeURIComponent(remoteClusterName)}/auth-groups?${params.toString()}`,
  )
    .then(handleResponse)
    .then((data) => {
      const response = data as LxdApiResponse<LxdAuthGroup[]>;
      response.metadata.forEach(ensureStableSorting);
      return response.metadata;
    });
};

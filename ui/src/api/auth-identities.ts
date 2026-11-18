import type { LxdApiResponse } from "types/apiResponse";
import type { LxdIdentity } from "types/permissions";
import { ensureStableSorting, handleResponse } from "util/helpers";

export const fetchIdentities = async (
  remoteClusterName: string,
): Promise<LxdIdentity[]> => {
  return fetch(
    `/1.0/remote-cluster/${encodeURIComponent(remoteClusterName)}/auth/identities`,
  )
    .then(handleResponse)
    .then((data) =>
      (data as LxdApiResponse<LxdIdentity[]>).metadata.map(ensureStableSorting),
    );
};

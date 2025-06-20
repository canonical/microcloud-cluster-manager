import { LxdApiResponse } from "types/apiResponse";
import { Token, TokenPostResponse } from "types/token";
import { handleResponse, handleSettledResult } from "util/helpers";

export const fetchTokens = (): Promise<Token[]> => {
  return new Promise((resolve, reject) => {
    fetch("/1.0/remote-cluster-join-token")
      .then(handleResponse)
      .then((data) => resolve((data as LxdApiResponse<Token[]>).metadata ?? []))
      .catch(reject);
  });
};

export const createToken = (body: string): Promise<TokenPostResponse> => {
  return new Promise((resolve, reject) => {
    fetch("/1.0/remote-cluster-join-token", {
      method: "POST",
      body: body,
    })
      .then(handleResponse)
      .then((data) =>
        resolve((data as LxdApiResponse<TokenPostResponse>).metadata),
      )
      .catch(reject);
  });
};

export const deleteToken = (remoteClusterName: string): Promise<void> => {
  return new Promise((resolve, reject) => {
    fetch(`/1.0/remote-cluster-join-token/${remoteClusterName}`, {
      method: "DELETE",
    })
      .then(handleResponse)
      .then(() => resolve())
      .catch(reject);
  });
};

export const deleteTokenBulk = async (
  remoteClusterNames: string[],
): Promise<void> => {
  return Promise.allSettled(
    remoteClusterNames.map(async (name) => deleteToken(name)),
  ).then(handleSettledResult);
};

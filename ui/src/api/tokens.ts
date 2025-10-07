import { LxdApiResponse } from "types/apiResponse";
import { Token, TokenPostResponse } from "types/token";
import { handleResponse, handleSettledResult } from "util/helpers";

export const fetchTokens = (): Promise<Token[]> => {
  return fetch("/1.0/remote-cluster-join-token")
    .then(handleResponse)
    .then((data) => (data as LxdApiResponse<Token[]>).metadata ?? []);
};

export const createToken = (body: string): Promise<TokenPostResponse> => {
  return fetch("/1.0/remote-cluster-join-token", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: body,
  })
    .then(handleResponse)
    .then((data) => (data as LxdApiResponse<TokenPostResponse>).metadata);
};

export const deleteToken = async (remoteClusterName: string): Promise<void> => {
  await fetch(`/1.0/remote-cluster-join-token/${remoteClusterName}`, {
    method: "DELETE",
  }).then(handleResponse);
};

export const deleteTokenBulk = async (
  remoteClusterNames: string[],
): Promise<void> => {
  return Promise.allSettled(
    remoteClusterNames.map(async (name) => deleteToken(name)),
  ).then(handleSettledResult);
};

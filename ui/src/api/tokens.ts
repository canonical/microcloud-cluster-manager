import { LxdApiResponse } from "types/apiResponse";
import { Token, TokenPostResponse } from "types/token";
import { handleResponse } from "util/helpers";

export const fetchTokens = (): Promise<Token[]> => {
  return new Promise((resolve, reject) => {
    fetch("/1.0/remote-cluster-join-token")
      .then(handleResponse)
      .then((data: LxdApiResponse<Token[]>) => resolve(data.metadata ?? []))
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
      .then((data: LxdApiResponse<TokenPostResponse>) => resolve(data.metadata))
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

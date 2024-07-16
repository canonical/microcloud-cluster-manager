import { LxdApiResponse } from "types/apiResponse";
import { Token } from "types/token";
import { handleResponse } from "util/helpers";

export const fetchTokens = (): Promise<Token[]> => {
  return new Promise((resolve, reject) => {
    fetch("/1.0/external-site-join-token")
      .then(handleResponse)
      .then((data: LxdApiResponse<Token[]>) => resolve(data.metadata ?? []))
      .catch(reject);
  });
};

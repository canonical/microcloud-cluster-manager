import type { LxdApiResponse } from "types/apiResponse";
import type { Server } from "types/server";
import { handleResponse } from "util/helpers";

export const fetchServer = async (): Promise<Server> => {
  return fetch("/1.0")
    .then(handleResponse)
    .then((data) => (data as LxdApiResponse<Server>).metadata);
};

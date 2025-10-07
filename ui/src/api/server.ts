import { LxdApiResponse } from "types/apiResponse";
import { Server } from "types/server";
import { handleResponse } from "util/helpers";

export const fetchServer = (): Promise<Server> => {
  return fetch("/1.0")
    .then(handleResponse)
    .then((data) => (data as LxdApiResponse<Server>).metadata);
};

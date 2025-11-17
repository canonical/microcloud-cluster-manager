import type { LxdApiResponse } from "types/apiResponse";
import type { Configuration } from "types/config";
import { handleResponse } from "util/helpers";

export const fetchConfigurations = async (): Promise<Configuration> => {
  return fetch("/1.0/configuration")
    .then(handleResponse)
    .then((data) => (data as LxdApiResponse<Configuration>).metadata);
};

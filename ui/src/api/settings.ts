import { LxdApiResponse } from "types/apiResponse";
import { Configuration } from "types/config";
import { handleResponse } from "util/helpers";

export const fetchConfigurations = (): Promise<Configuration> => {
  return fetch("/1.0/configuration")
    .then(handleResponse)
    .then((data) => (data as LxdApiResponse<Configuration>).metadata);
};

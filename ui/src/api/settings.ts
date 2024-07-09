import { LxdApiResponse } from "types/apiResponse";
import { LxdConfigPair, ManagerOptions } from "types/config";
import { handleResponse } from "util/helpers";

export const fetchConfigOptions = (): Promise<ManagerOptions> => {
  return new Promise((resolve, reject) => {
    fetch("/1.0/config")
      .then(handleResponse)
      .then((data: LxdApiResponse<ManagerOptions>) => resolve(data.metadata))
      .catch(reject);
  });
};

export const updateSettings = (config: LxdConfigPair): Promise<void> => {
  return new Promise((resolve, reject) => {
    fetch("/1.0/config", {
      method: "PATCH",
      body: JSON.stringify({
        config,
      }),
    })
      .then(handleResponse)
      .then(resolve)
      .catch(reject);
  });
};

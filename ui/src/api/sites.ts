import { LxdApiResponse } from "types/apiResponse";
import { Site } from "types/site";
import { handleResponse } from "util/helpers";

export const fetchSites = (): Promise<Site[]> => {
  return new Promise((resolve, reject) => {
    fetch("/1.0/sites")
      .then(handleResponse)
      .then((data: LxdApiResponse<Site[]>) => resolve(data.metadata))
      .catch(reject);
  });
};

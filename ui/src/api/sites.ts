import { LxdApiResponse } from "types/apiResponse";
import { LxdSite } from "types/site";
import { handleResponse } from "util/helpers";

export const fetchSites = (): Promise<LxdSite[]> => {
  return new Promise((resolve, reject) => {
    fetch("/1.0/sites")
      .then(handleResponse)
      .then((data: LxdApiResponse<LxdSite[]>) => resolve(data.metadata))
      .catch(reject);
  });
};

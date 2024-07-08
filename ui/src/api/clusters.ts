import { LxdApiResponse } from "types/apiResponse";
import { Cluster } from "types/cluster";
import { handleResponse } from "util/helpers";

export const fetchClusters = (): Promise<Cluster[]> => {
  return new Promise((resolve, reject) => {
    fetch("/1.0/sites")
      .then(handleResponse)
      .then((data: LxdApiResponse<Cluster[]>) => resolve(data.metadata))
      .catch(reject);
  });
};

export const deleteCluster = (siteName: string): Promise<void> => {
  return new Promise((resolve, reject) => {
    fetch(`/1.0/sites/${siteName}`, {
      method: "DELETE",
    })
      .then(handleResponse)
      .then(() => resolve())
      .catch(reject);
  });
};

import { LxdApiResponse } from "types/apiResponse";
import { Cluster } from "types/cluster";
import { handleResponse } from "util/helpers";

export const fetchClusters = (): Promise<Cluster[]> => {
  return new Promise((resolve, reject) => {
    fetch("/1.0/remote-clusters")
      .then(handleResponse)
      .then((data: LxdApiResponse<Cluster[]>) => resolve(data.metadata))
      .catch(reject);
  });
};

export const deleteCluster = (remoteClusterName: string): Promise<void> => {
  return new Promise((resolve, reject) => {
    fetch(`/1.0/remote-clusters/${remoteClusterName}`, {
      method: "DELETE",
    })
      .then(handleResponse)
      .then(() => resolve())
      .catch(reject);
  });
};

export const approveCluster = (remoteClusterName: string): Promise<void> => {
  return new Promise((resolve, reject) => {
    fetch(`/1.0/remote-clusters/${remoteClusterName}`, {
      method: "PATCH",
      body: JSON.stringify({ status: "ACTIVE" }),
    })
      .then(handleResponse)
      .then(resolve)
      .catch(reject);
  });
};

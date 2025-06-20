import { LxdApiResponse } from "types/apiResponse";
import { Cluster } from "types/cluster";
import { handleResponse, handleSettledResult } from "util/helpers";

export const fetchClusters = (): Promise<Cluster[]> => {
  return new Promise((resolve, reject) => {
    fetch("/1.0/remote-cluster")
      .then(handleResponse)
      .then((data) =>
        resolve((data as LxdApiResponse<Cluster[]>).metadata ?? []),
      )
      .catch(reject);
  });
};

export const fetchCluster = (remoteClusterName: string): Promise<Cluster> => {
  return new Promise((resolve, reject) => {
    fetch(`/1.0/remote-cluster/${remoteClusterName}`)
      .then(handleResponse)
      .then((data) => resolve((data as LxdApiResponse<Cluster>).metadata))
      .catch(reject);
  });
};

export const deleteCluster = (remoteClusterName: string): Promise<void> => {
  return new Promise((resolve, reject) => {
    fetch(`/1.0/remote-cluster/${remoteClusterName}`, {
      method: "DELETE",
    })
      .then(handleResponse)
      .then(() => resolve())
      .catch(reject);
  });
};

export const deleteClusterBulk = async (
  remoteClusterNames: string[],
): Promise<void> => {
  return Promise.allSettled(
    remoteClusterNames.map(async (name) => deleteCluster(name)),
  ).then(handleSettledResult);
};

export const approveCluster = (remoteClusterName: string): Promise<void> => {
  return new Promise((resolve, reject) => {
    fetch(`/1.0/remote-cluster/${remoteClusterName}`, {
      method: "PATCH",
      body: JSON.stringify({ status: "ACTIVE" }),
    })
      .then(handleResponse)
      .then(() => resolve())
      .catch(reject);
  });
};

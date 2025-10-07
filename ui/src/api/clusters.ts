import { LxdApiResponse } from "types/apiResponse";
import { Cluster } from "types/cluster";
import { handleResponse, handleSettledResult } from "util/helpers";

export const fetchClusters = (): Promise<Cluster[]> => {
  return fetch("/1.0/remote-cluster")
    .then(handleResponse)
    .then((data) => (data as LxdApiResponse<Cluster[]>).metadata ?? []);
};

export const fetchCluster = (remoteClusterName: string): Promise<Cluster> => {
  return fetch(`/1.0/remote-cluster/${remoteClusterName}`)
    .then(handleResponse)
    .then((data) => (data as LxdApiResponse<Cluster>).metadata);
};

export const deleteCluster = async (
  remoteClusterName: string,
): Promise<void> => {
  await fetch(`/1.0/remote-cluster/${remoteClusterName}`, {
    method: "DELETE",
  }).then(handleResponse);
};

export const deleteClusterBulk = async (
  remoteClusterNames: string[],
): Promise<void> => {
  return Promise.allSettled(
    remoteClusterNames.map(async (name) => deleteCluster(name)),
  ).then(handleSettledResult);
};

export const updateCluster = async (
  remoteClusterName: string,
  payload: string,
): Promise<void> => {
  await fetch(`/1.0/remote-cluster/${remoteClusterName}`, {
    method: "PATCH",
    headers: {
      "Content-Type": "application/json",
    },
    body: payload,
  }).then(handleResponse);
};

export const updateClusterBulk = async (
  remoteClusterNames: string[],
  payload: string,
): Promise<void> => {
  return Promise.allSettled(
    remoteClusterNames.map(async (name) => updateCluster(name, payload)),
  ).then(handleSettledResult);
};

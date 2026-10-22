import type { LxdApiResponse } from "types/apiResponse";
import type { Cluster, ClusterLink } from "types/cluster";
import { handleResponse, handleSettledResult } from "util/helpers";

export const fetchClusters = async (): Promise<Cluster[]> => {
  return fetch("/1.0/remote-cluster")
    .then(handleResponse)
    .then((data) => (data as LxdApiResponse<Cluster[]>).metadata ?? []);
};

export const fetchCluster = async (
  remoteClusterName: string,
): Promise<Cluster> => {
  return fetch(`/1.0/remote-cluster/${encodeURIComponent(remoteClusterName)}`)
    .then(handleResponse)
    .then((data) => (data as LxdApiResponse<Cluster>).metadata);
};

export const deleteCluster = async (
  remoteClusterName: string,
): Promise<void> => {
  await fetch(`/1.0/remote-cluster/${encodeURIComponent(remoteClusterName)}`, {
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
  await fetch(`/1.0/remote-cluster/${encodeURIComponent(remoteClusterName)}`, {
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

export const fetchClusterLinks = async (
  remoteClusterName: string,
): Promise<ClusterLink[]> => {
  return fetch(
    `/1.0/remote-cluster/${encodeURIComponent(remoteClusterName)}/cluster-links`,
  )
    .then(handleResponse)
    .then((data) => (data as LxdApiResponse<ClusterLink[]>).metadata);
};

export const createClusterLink = async (
  source: string,
  payload: string,
): Promise<void> => {
  await fetch(
    `/1.0/remote-cluster/${encodeURIComponent(source)}/cluster-links`,
    {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: payload,
    },
  ).then(handleResponse);
};

export const deleteClusterLink = async (
  source: string,
  clusterLinkName: string,
): Promise<void> => {
  await fetch(
    `/1.0/remote-cluster/${encodeURIComponent(source)}/cluster-links/${encodeURIComponent(clusterLinkName)}`,
    {
      method: "DELETE",
    },
  ).then(handleResponse);
};

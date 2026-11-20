import type { LxdApiResponse } from "types/apiResponse";
import type {
  Cluster,
  ClusterLinkList,
  ClusterLinkState,
  ResolvedClusterLink,
} from "types/cluster";
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
  cluster: string,
): Promise<ResolvedClusterLink[]> => {
  const params = new URLSearchParams();
  params.set("cluster", cluster);

  const { cluster_links, errors } = await fetch(
    `/1.0/cluster-links?${params.toString()}`,
  )
    .then(handleResponse)
    .then((data) => (data as LxdApiResponse<ClusterLinkList>).metadata);
  if (errors.length > 0) {
    throw new Error(errors[0].error);
  }
  return cluster_links;
};

export const createClusterLink = async (
  source: string,
  payload: string,
): Promise<void> => {
  const params = new URLSearchParams();
  params.set("cluster", source);

  await fetch(`/1.0/cluster-links?${params.toString()}`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: payload,
  }).then(handleResponse);
};

export const updateClusterLink = async (
  source: string,
  clusterLinkName: string,
  body: string,
): Promise<void> => {
  const params = new URLSearchParams();
  params.set("cluster", source);

  await fetch(
    `/1.0/cluster-links/${encodeURIComponent(clusterLinkName)}?${params.toString()}`,
    {
      method: "PUT",
      headers: {
        "Content-Type": "application/json",
      },
      body,
    },
  ).then(handleResponse);
};

export const deleteClusterLink = async (
  source: string,
  clusterLinkName: string,
  cleanupTarget: boolean,
): Promise<void> => {
  const params = new URLSearchParams();
  params.set("cluster", source);
  if (cleanupTarget) {
    params.set("cleanupTarget", "true");
  }

  await fetch(
    `/1.0/cluster-links/${encodeURIComponent(clusterLinkName)}?${params.toString()}`,
    {
      method: "DELETE",
    },
  ).then(handleResponse);
};

export const fetchClusterLinkState = async (
  remoteClusterName: string,
  linkName: string,
): Promise<ClusterLinkState> => {
  return fetch(
    `/1.0/remote-cluster/${encodeURIComponent(
      remoteClusterName,
    )}/cluster-links/${encodeURIComponent(linkName)}/state`,
  )
    .then(handleResponse)
    .then((data) => (data as LxdApiResponse<ClusterLinkState>).metadata);
};

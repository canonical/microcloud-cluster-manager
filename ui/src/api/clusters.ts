import type { LxdApiResponse } from "types/apiResponse";
import type {
  Cluster,
  ClusterLinkList,
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

const clusterLinksUrl = (
  cluster: string,
  clusterLinkName?: string,
  params: Record<string, string> = {},
): string => {
  const path = clusterLinkName
    ? `/1.0/cluster-links/${encodeURIComponent(clusterLinkName)}`
    : "/1.0/cluster-links";
  const query = new URLSearchParams({ cluster, ...params });
  return `${path}?${query.toString()}`;
};

export const fetchClusterLinks = async (
  cluster: string,
): Promise<ResolvedClusterLink[]> => {
  return fetch(clusterLinksUrl(cluster))
    .then(handleResponse)
    .then(
      (data) =>
        (data as LxdApiResponse<ClusterLinkList>).metadata.cluster_links,
    );
};

export const createClusterLink = async (
  source: string,
  payload: string,
): Promise<void> => {
  await fetch(clusterLinksUrl(source), {
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
  await fetch(clusterLinksUrl(source, clusterLinkName), {
    method: "PUT",
    headers: {
      "Content-Type": "application/json",
    },
    body,
  }).then(handleResponse);
};

export const deleteClusterLink = async (
  source: string,
  clusterLinkName: string,
  removeReverseEdge: boolean,
): Promise<void> => {
  const params: Record<string, string> = removeReverseEdge
    ? { removeReverseEdge: "true" }
    : {};

  await fetch(clusterLinksUrl(source, clusterLinkName, params), {
    method: "DELETE",
  }).then(handleResponse);
};

export const deleteClusterLinks = async (
  source: string,
  links: { name: string; removeReverseEdge: boolean }[],
): Promise<void> => {
  return Promise.allSettled(
    links.map(async ({ name, removeReverseEdge }) =>
      deleteClusterLink(source, name, removeReverseEdge),
    ),
  ).then(handleSettledResult);
};

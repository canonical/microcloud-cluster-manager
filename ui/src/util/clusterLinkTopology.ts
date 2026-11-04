import type { Cluster, ClusterLink, ClusterLinkType } from "types/cluster";
import { getLinkedCluster } from "util/clusterLinks";

export const hasOnwardClusterLinks = (
  links: ClusterLink[],
  sourceCluster: Cluster,
): boolean => {
  return links.some((link) => !getLinkedCluster(link, [sourceCluster]));
};

export interface ClusterLinkTopologyItem {
  clusterLink: ClusterLink;
  id: string;
  name: string;
  type: ClusterLinkType;
  targetClusterName?: string;
  hasMore?: boolean;
}

export const toClusterLinkTopologyItem = (
  link: ClusterLink,
  clusters: Cluster[],
): ClusterLinkTopologyItem => {
  if (
    link.type !== "bidirectional" &&
    link.type !== "unidirectional" &&
    link.type !== "public"
  ) {
    throw new Error(
      `Unsupported cluster link type "${link.type as string}" for "${link.name}".`,
    );
  }

  const targetCluster = getLinkedCluster(link, clusters);

  return {
    clusterLink: link,
    id: link.name,
    name: link.name,
    type: link.type,
    targetClusterName: targetCluster?.name,
    hasMore: targetCluster ? undefined : true,
  };
};

import type { Cluster, ClusterLink } from "types/cluster";

const getLinkAddresses = (link: ClusterLink): string[] => {
  return link.config?.["volatile.addresses"]?.split(",") ?? [];
};

// Resolves the cluster a link points at by UUID, falling back to matching its
// addresses against the lxd URLs of clusters known to the cluster manager.
export const getLinkedCluster = (
  link: ClusterLink,
  clusters: Cluster[],
): Cluster | undefined => {
  const linkedCluster = clusters.find(
    (cluster) => cluster.cluster_uuid === link.config?.["volatile.uuid"],
  );

  if (linkedCluster) {
    return linkedCluster;
  }

  // fallback to address matching if no UUID match is found.
  const addresses = getLinkAddresses(link);
  return clusters.find((cluster) =>
    addresses.some((address) => cluster.lxd_url.endsWith(address)),
  );
};

// Finds the link on the other end of a cluster link, meaning the link on the
// linked cluster that points back at the given cluster.
export const getReverseLink = (
  links: ClusterLink[],
  cluster: Cluster,
): ClusterLink | undefined => {
  return links.find((link) => {
    if (link.config?.["volatile.uuid"] === cluster.cluster_uuid) {
      return true;
    }

    return getLinkAddresses(link).some((address) =>
      cluster.lxd_url.endsWith(address),
    );
  });
};

import type { Cluster, ClusterLink } from "types/cluster";

const getLinkAddresses = (link: ClusterLink): string[] => {
  return link.config?.["volatile.addresses"]?.split(",") ?? [];
};

// Resolves the cluster a link points at, by matching the addresses of the link
// against the ui url of the clusters known to the cluster manager.
export const getLinkedCluster = (
  link: ClusterLink,
  clusters: Cluster[],
): Cluster | undefined => {
  const addresses = getLinkAddresses(link);

  return clusters.find((cluster) =>
    addresses.some((address) => `https://${address}` === cluster.ui_url),
  );
};

// Finds the link on the other end of a cluster link, meaning the link on the
// linked cluster that points back at the given cluster.
export const getReverseLink = (
  links: ClusterLink[],
  cluster: Cluster,
): ClusterLink | undefined => {
  return links.find((link) =>
    getLinkAddresses(link).some(
      (address) => `https://${address}` === cluster.ui_url,
    ),
  );
};

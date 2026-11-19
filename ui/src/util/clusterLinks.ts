import type { Cluster, ClusterLink } from "types/cluster";
import type { LxdIdentity } from "types/permissions";

export const getLinkAddresses = (link: ClusterLink): string[] => {
  return (
    link.config?.["volatile.addresses"]
      ?.split(",")
      .map((address) => address.trim())
      .filter(Boolean) ?? []
  );
};

// Addresses to show for a link. Falls back to the pending address while the
// link has no resolved addresses yet.
export const getLinkDisplayAddresses = (link: ClusterLink): string[] => {
  const addresses = getLinkAddresses(link);
  if (addresses.length) {
    return addresses;
  }

  const pendingAddress = link.config?.["volatile.pending_address"];
  return pendingAddress ? [pendingAddress] : [];
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

export const isBidirectional = (link: ClusterLink): boolean =>
  link.type.toLowerCase() === "bidirectional";

export const isUnidirectional = (link: ClusterLink): boolean =>
  link.type.toLowerCase() === "unidirectional";

export const isPublic = (link: ClusterLink): boolean =>
  link.type.toLowerCase() === "public";

// used_by holds LXD API paths, e.g. /1.0/replicators/<name>
export const getLinkUsage = (
  link: ClusterLink,
): { replicators: number; imageRegistries: number } => {
  const usedBy = link.used_by ?? [];
  const countPrefix = (prefix: string) =>
    usedBy.filter((path) => path.startsWith(prefix)).length;

  return {
    replicators: countPrefix("/1.0/replicators/"),
    imageRegistries: countPrefix("/1.0/image-registries/"),
  };
};

// Bidirectional and unidirectional links have an end on the target cluster.
export const hasTargetEnd = (link: ClusterLink): boolean =>
  isBidirectional(link) || isUnidirectional(link);

// The target end can only be cleaned up when the target has a registered tunnel.
export const canRemoveTargetEnd = (
  link: ClusterLink,
  targetCluster?: Cluster,
): boolean => hasTargetEnd(link) && !!targetCluster?.tunnel_registered;

export const getLinkIdentity = (
  identities: LxdIdentity[],
  linkName: string | undefined | null,
) => {
  if (!linkName) return undefined;

  return identities.find(
    (identity) =>
      identity.name === linkName &&
      identity.type.startsWith("Cluster link certificate"),
  );
};

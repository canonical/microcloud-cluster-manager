import type { LxdIdentity } from "types/permissions";
import type { Cluster, ClusterLink } from "types/cluster";

export const getLinkAddresses = (link: ClusterLink): string[] => {
  const addresses =
    link.config?.["volatile.addresses"]
      ?.split(",")
      .map((address) => address.trim())
      .filter(Boolean) ?? [];
  const pendingAddress = link.config?.["volatile.pending_address"]?.trim();

  return addresses.length || !pendingAddress ? addresses : [pendingAddress];
};

// A connection over a cluster link: the link itself, the source cluster it is
// defined on, and the target cluster it points to, if the target is known to
// the cluster manager.
export interface ClusterConnection {
  id: string;
  source: Cluster;
  link: ClusterLink;
  target?: Cluster;
}

export interface ClusterLinkFilter {
  queries: string[];
  types: string[];
}

export const filterClusterLinks = (
  entries: ClusterConnection[],
  filters: ClusterLinkFilter,
): ClusterConnection[] => {
  const queries = filters.queries.map((query) => query.toLowerCase());
  return entries.filter(
    ({ source, link, target }) =>
      (!queries.length ||
        queries.some((query) =>
          [
            source.name,
            link.name,
            link.description ?? "",
            target?.name ?? "",
            ...getLinkAddresses(link),
          ].some((value) => value.toLowerCase().includes(query)),
        )) &&
      (!filters.types.length ||
        filters.types.includes(link.type.toLowerCase())),
  );
};

export const getClusterLinkId = (
  sourceName: string,
  linkName: string,
): string => `${sourceName}/${linkName}`;

export const getSourceClusterConnections = (
  source: Cluster,
  links: ClusterLink[],
  clusters: Cluster[],
): ClusterConnection[] =>
  links.map((link) => ({
    id: getClusterLinkId(source.name, link.name),
    source,
    link,
    target: getLinkedCluster(link, clusters),
  }));

export const getClusterLinkSortData = (entry: ClusterConnection) => ({
  name: entry.link.name.toLowerCase(),
  type: entry.link.type.toLowerCase(),
  target: (entry.target?.name ?? "").toLowerCase(),
  description: (entry.link.description ?? "").toLowerCase(),
  usedBy: entry.link.used_by?.length ?? 0,
  addresses: getLinkAddresses(entry.link).join(",").toLowerCase(),
});

// Returns the host and port of the lxd URL of a cluster, in the format of the
// addresses of a cluster link.
const getClusterAddress = (cluster: Cluster): string | undefined => {
  try {
    return new URL(cluster.lxd_url).host.toLowerCase();
  } catch {
    return undefined;
  }
};

const hasLinkAddress = (cluster: Cluster, link: ClusterLink): boolean => {
  const clusterAddress = getClusterAddress(cluster);
  return (
    !!clusterAddress &&
    getLinkAddresses(link).some(
      (address) => address.toLowerCase() === clusterAddress,
    )
  );
};

// Resolves the cluster a link points at by UUID, falling back to matching its
// confirmed or pending addresses against the lxd URLs of known clusters.
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
  return clusters.find((cluster) => hasLinkAddress(cluster, link));
};

// Finds the bidirectional link on the other end of a cluster link, meaning the
// link on the linked cluster that points back at the given cluster. Other link
// types have no counterpart on the linked cluster.
export const getReverseLink = (
  links: ClusterLink[],
  cluster: Cluster,
): ClusterLink | undefined => {
  return links.find((link) => {
    if (link.type.toLowerCase() !== "bidirectional") {
      return false;
    }

    return (
      link.config?.["volatile.uuid"] === cluster.cluster_uuid ||
      hasLinkAddress(cluster, link)
    );
  });
};

export const isBidirectional = (link: ClusterLink): boolean =>
  link.type.toLowerCase() === "bidirectional";

export const isUnidirectional = (link: ClusterLink): boolean =>
  link.type.toLowerCase() === "unidirectional";

export const isPublic = (link: ClusterLink): boolean =>
  link.type.toLowerCase() === "public";

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

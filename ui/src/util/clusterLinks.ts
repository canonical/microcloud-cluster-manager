import type { Cluster, ClusterLink } from "types/cluster";

export const linkTypeIcons: Record<string, string | undefined> = {
  bidirectional: "swap",
  unidirectional: "arrow-right",
  public: "exposed",
};

export const getLinkAddresses = (link: ClusterLink): string[] => {
  return (
    link.config?.["volatile.addresses"]
      ?.split(",")
      .map((address) => address.trim())
      .filter(Boolean) ?? []
  );
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

export type ClusterLinkSortKey =
  | "cluster"
  | "name"
  | "type"
  | "target"
  | "description"
  | "usedBy"
  | "addresses";

export type ClusterLinkSortDirection = "ascending" | "descending";

export const filterClusterLinks = (
  entries: ClusterConnection[],
  filters: ClusterLinkFilter,
): ClusterConnection[] => {
  const queries = filters.queries.map((query) => query.toLowerCase());
  return entries.filter(
    ({ link, target }) =>
      (!queries.length ||
        queries.some((query) =>
          [
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

// Flattens per-cluster link query results into entries in cluster order, so
// grouping by source cluster later preserves the order of the cluster list.
export const flattenClusterLinks = (
  sourceClusters: Cluster[],
  linksByCluster: Record<string, ClusterLink[] | undefined>,
  enrolledClusters: Cluster[] = sourceClusters,
): ClusterConnection[] => {
  const entries: ClusterConnection[] = [];
  for (const source of sourceClusters) {
    for (const link of linksByCluster[source.name] ?? []) {
      entries.push({
        id: getClusterLinkId(source.name, link.name),
        source,
        link,
        target: getLinkedCluster(link, enrolledClusters),
      });
    }
  }
  return entries;
};

const getSortValue = (
  entry: ClusterConnection,
  sortKey: Exclude<ClusterLinkSortKey, "cluster">,
): string | number => {
  switch (sortKey) {
    case "name":
      return entry.link.name.toLowerCase();
    case "type":
      return entry.link.type.toLowerCase();
    case "target":
      return (entry.target?.name ?? "").toLowerCase();
    case "description":
      return (entry.link.description ?? "").toLowerCase();
    case "usedBy":
      return entry.link.used_by?.length ?? 0;
    case "addresses":
      return getLinkAddresses(entry.link).join(",").toLowerCase();
  }
};

export interface SourceClusterGroup {
  cluster: Cluster;
  links: ClusterConnection[];
}

// Groups entries by their source cluster, keeping the cluster order of the
// flattened input. Groups without entries are omitted.
export const groupClusterLinks = (
  entries: ClusterConnection[],
): SourceClusterGroup[] => {
  const groupsByCluster = new Map<string, SourceClusterGroup>();
  for (const entry of entries) {
    const group = groupsByCluster.get(entry.source.name);
    if (group) {
      group.links.push(entry);
    } else {
      groupsByCluster.set(entry.source.name, {
        cluster: entry.source,
        links: [entry],
      });
    }
  }
  return [...groupsByCluster.values()];
};

export const sortClusterLinks = (
  entries: ClusterConnection[],
  sortKey: ClusterLinkSortKey | null,
  sortDirection: ClusterLinkSortDirection = "ascending",
  groupBySource = false,
): ClusterConnection[] => {
  if (!sortKey) {
    return [...entries];
  }

  const direction = sortDirection === "ascending" ? 1 : -1;
  if (sortKey === "cluster") {
    return groupClusterLinks(entries)
      .sort((a, b) => a.cluster.name.localeCompare(b.cluster.name) * direction)
      .flatMap((group) => group.links);
  }

  const compare = (a: ClusterConnection, b: ClusterConnection): number => {
    const valueA = getSortValue(a, sortKey);
    const valueB = getSortValue(b, sortKey);
    if (valueA < valueB) {
      return -1 * direction;
    }
    if (valueA > valueB) {
      return direction;
    }
    return a.id.localeCompare(b.id);
  };

  return groupBySource
    ? groupClusterLinks(entries).flatMap((group) => group.links.sort(compare))
    : [...entries].sort(compare);
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

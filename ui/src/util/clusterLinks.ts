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

// Sort data attached to the MainTable rows of the cluster link list. All rows
// of a group share the group index (its position in the unsorted list) and the
// cluster name. Group header rows are flagged with isGroup.
export interface ClusterLinkRowSortData extends Record<string, unknown> {
  group: number;
  isGroup: 0 | 1;
  cluster: string;
  id: string;
}

// Builds the sort data for a group header row, keeping the header attached to
// the top of its group when sorting.
export const getClusterLinkGroupSortData = (
  clusterName: string,
  group: number,
): ClusterLinkRowSortData => ({
  group,
  isGroup: 1,
  cluster: clusterName,
  id: `group-${clusterName}`,
});

// Builds the sort data for a link row, so sorting compares the same values
// that are displayed in the columns.
export const getClusterLinkSortData = (
  entry: ClusterConnection,
  group: number,
): ClusterLinkRowSortData => ({
  group,
  isGroup: 0,
  cluster: entry.source.name,
  id: entry.id,
  name: entry.link.name.toLowerCase(),
  type: entry.link.type.toLowerCase(),
  target: (entry.target?.name ?? "").toLowerCase(),
  description: (entry.link.description ?? "").toLowerCase(),
  usedBy: entry.link.used_by?.length ?? 0,
  addresses: getLinkAddresses(entry.link).join(",").toLowerCase(),
});

// Compares two cluster link list rows for MainTable's sortFunction. Groups
// stay contiguous with the group header row first and link rows are sorted
// within their group only. Sorting by cluster reorders the groups themselves
// and leaves the link order within groups untouched.
export const compareClusterLinkRows = (
  a: ClusterLinkRowSortData,
  b: ClusterLinkRowSortData,
  key: string,
  direction: "ascending" | "descending",
): -1 | 0 | 1 => {
  const sign = direction === "ascending" ? 1 : -1;

  if (a.cluster !== b.cluster) {
    if (key === "cluster") {
      return (Math.sign(a.cluster.localeCompare(b.cluster)) * sign) as
        | -1
        | 0
        | 1;
    }
    return a.group < b.group ? -1 : 1;
  }

  if (a.isGroup !== b.isGroup) {
    return a.isGroup === 1 ? -1 : 1;
  }

  if (key === "cluster") {
    return 0;
  }

  const valueA = a[key] as string | number;
  const valueB = b[key] as string | number;
  if (valueA === valueB) {
    if (a.id === b.id) {
      return 0;
    }
    return a.id < b.id ? -1 : 1;
  }
  return ((valueA < valueB ? -1 : 1) * sign) as -1 | 0 | 1;
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

import type { SortDirection } from "@canonical/react-components";
import type {
  MainTableHeader,
  MainTableRow,
} from "@canonical/react-components/dist/components/MainTable/MainTable";
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

// Builds the connections of the clusters in cluster order, so grouping by
// source cluster later preserves the order of the cluster list. The target of a
// connection is resolved among the clusters known to the cluster manager.
export const getClusterConnections = (
  clusters: Cluster[],
): ClusterConnection[] =>
  clusters.flatMap((source) =>
    (source.cluster_links ?? []).map((link) => ({
      id: getClusterLinkId(source.name, link.name),
      source,
      link,
      target: getLinkedCluster(link, clusters),
    })),
  );

export interface GlobalClusterLinkSourceGroup {
  cluster: Cluster;
  links: ClusterConnection[];
}

// Groups entries by their source cluster, keeping the cluster order of the
// flattened input. Groups without entries are omitted.
export const groupGlobalClusterLinks = (
  entries: ClusterConnection[],
): GlobalClusterLinkSourceGroup[] => {
  const groupsByCluster = new Map<string, GlobalClusterLinkSourceGroup>();
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
export interface GlobalClusterLinkRowSortData extends Record<string, unknown> {
  group: number;
  isGroup: 0 | 1;
  cluster: string;
  id: string;
}

// Builds the sort data for a group header row, keeping the header attached to
// the top of its group when sorting.
export const getGlobalClusterLinkGroupSortData = (
  clusterName: string,
  group: number,
): GlobalClusterLinkRowSortData => ({
  group,
  isGroup: 1,
  cluster: clusterName,
  id: `group-${clusterName}`,
});

// Builds the sort data for a link row, so sorting compares the same values
// that are displayed in the columns.
export const getGlobalClusterLinkSortData = (
  entry: ClusterConnection,
  group: number,
): GlobalClusterLinkRowSortData => ({
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

// Compares two cluster link list rows. Groups
// stay contiguous with the group header row first and link rows are sorted
// within their group only. Sorting by cluster reorders the groups themselves
// and leaves the link order within groups untouched.
export const compareGlobalClusterLinkRows = (
  a: GlobalClusterLinkRowSortData,
  b: GlobalClusterLinkRowSortData,
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

// MainTable sort function that sorts link rows within their group only,
// keeping groups contiguous with the group header row first.
export const sortGlobalClusterLinkRows = (
  a: MainTableRow,
  b: MainTableRow,
  direction: SortDirection,
  key: MainTableHeader["sortKey"],
): -1 | 0 | 1 =>
  compareGlobalClusterLinkRows(
    a.sortData as GlobalClusterLinkRowSortData,
    b.sortData as GlobalClusterLinkRowSortData,
    String(key),
    direction === "descending" ? "descending" : "ascending",
  );

// Sorts the connections like the cluster link table sorts its rows: link
// rows are sorted within their source cluster, and sorting by cluster
// reorders the clusters. Sorting all connections before they are paginated
// makes the sort order apply across pages.
export const sortGlobalClusterConnections = (
  connections: ClusterConnection[],
  key: string | null | undefined,
  direction: "ascending" | "descending",
): ClusterConnection[] => {
  if (!key) {
    return connections;
  }

  const groupIndexes = new Map<string, number>();
  for (const { source } of connections) {
    if (!groupIndexes.has(source.name)) {
      groupIndexes.set(source.name, groupIndexes.size);
    }
  }

  const sortData = new Map<string, GlobalClusterLinkRowSortData>();
  for (const connection of connections) {
    sortData.set(
      connection.id,
      getGlobalClusterLinkSortData(
        connection,
        groupIndexes.get(connection.source.name) ?? 0,
      ),
    );
  }

  return [...connections].sort((a, b) => {
    const sortDataA = sortData.get(a.id);
    const sortDataB = sortData.get(b.id);
    if (!sortDataA || !sortDataB) {
      return 0;
    }
    return compareGlobalClusterLinkRows(sortDataA, sortDataB, key, direction);
  });
};

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

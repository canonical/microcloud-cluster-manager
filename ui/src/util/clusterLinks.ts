import type { QueryClient } from "@tanstack/react-query";
import type { SortDirection } from "@canonical/react-components";
import type {
  MainTableHeader,
  MainTableRow,
} from "@canonical/react-components/dist/components/MainTable/MainTable";
import type {
  Cluster,
  ClusterLink,
  ClusterLinkMemberState,
  ClusterLinkState,
  ResolvedClusterLink,
  StatusCaption,
} from "types/cluster";
import type { LxdIdentity } from "types/permissions";
import { allClusterLinksQueryKey, queryKeys } from "util/queryKeys";
import { compareGroupedRows } from "util/groupedTableSort";
import type { GroupedRowSortData } from "util/groupedTableSort";

export const getLinkAddresses = (link: ClusterLink): string[] => {
  return (
    link.config?.["volatile.addresses"]
      ?.split(",")
      .map((address) => address.trim())
      .filter(Boolean) ?? []
  );
};

// Pending addresses are shown before the link has confirmed addresses.
export const getLinkDisplayAddresses = (link: ClusterLink): string[] => {
  const addresses = getLinkAddresses(link);
  const pendingAddress = link.config?.["volatile.pending_address"]?.trim();

  return addresses.length || !pendingAddress ? addresses : [pendingAddress];
};

export interface ClusterLinkFilter {
  queries: string[];
  types: string[];
}

export const filterClusterLinks = (
  links: ResolvedClusterLink[],
  filters: ClusterLinkFilter,
): ResolvedClusterLink[] => {
  const queries = filters.queries.map((query) => query.toLowerCase());
  return links.filter(
    ({ source_cluster, link, target_cluster }) =>
      (!queries.length ||
        queries.some((query) =>
          [
            source_cluster,
            link.name,
            link.description ?? "",
            target_cluster ?? "",
            ...getLinkDisplayAddresses(link),
          ].some((value) => value.toLowerCase().includes(query)),
        )) &&
      (!filters.types.length ||
        filters.types.includes(link.type.toLowerCase())),
  );
};

export const getClusterLinkId = ({
  source_cluster,
  link,
}: ResolvedClusterLink): string => `${source_cluster}/${link.name}`;

export interface GlobalClusterLinkSourceGroup {
  sourceCluster: string;
  links: ResolvedClusterLink[];
}

// Groups links by their source cluster, keeping the cluster order of the
// flattened input. Groups without links are omitted.
export const groupGlobalClusterLinks = (
  links: ResolvedClusterLink[],
): GlobalClusterLinkSourceGroup[] => {
  const groupsByCluster = new Map<string, GlobalClusterLinkSourceGroup>();
  for (const resolvedLink of links) {
    const group = groupsByCluster.get(resolvedLink.source_cluster);
    if (group) {
      group.links.push(resolvedLink);
    } else {
      groupsByCluster.set(resolvedLink.source_cluster, {
        sourceCluster: resolvedLink.source_cluster,
        links: [resolvedLink],
      });
    }
  }
  return [...groupsByCluster.values()];
};

export const CLUSTER_LINK_GROUP_SORT_COLUMN = "cluster";

export const getGlobalClusterLinkGroupSortData = (
  groupKey: string,
  groupIndex: number,
): GroupedRowSortData => ({
  groupIndex,
  groupKey,
  isGroupHeader: true,
  id: `group-${groupKey}`,
  values: {},
});

export const getClusterLinkSortData = ({
  link,
  target_cluster,
}: ResolvedClusterLink) => ({
  name: link.name.toLowerCase(),
  type: link.type.toLowerCase(),
  target: (target_cluster ?? "").toLowerCase(),
  description: (link.description ?? "").toLowerCase(),
  usedBy: link.used_by?.length ?? 0,
  addresses: getLinkDisplayAddresses(link).join(",").toLowerCase(),
});

export const isBidirectional = (link: ClusterLink): boolean =>
  link.type.toLowerCase() === "bidirectional";

export const isUnidirectional = (link: ClusterLink): boolean =>
  link.type.toLowerCase() === "unidirectional";

export const isPublic = (link: ClusterLink): boolean =>
  link.type.toLowerCase() === "public";

export const getGlobalClusterLinkSortData = (
  resolvedLink: ResolvedClusterLink,
  groupIndex: number,
): GroupedRowSortData => ({
  groupIndex,
  groupKey: resolvedLink.source_cluster,
  isGroupHeader: false,
  id: getClusterLinkId(resolvedLink),
  values: getClusterLinkSortData(resolvedLink),
});

export const sortGlobalClusterLinkRows = (
  a: MainTableRow,
  b: MainTableRow,
  direction: SortDirection,
  key: MainTableHeader["sortKey"],
): -1 | 0 | 1 =>
  compareGroupedRows(
    a.sortData as GroupedRowSortData,
    b.sortData as GroupedRowSortData,
    String(key),
    direction === "descending" ? "descending" : "ascending",
    CLUSTER_LINK_GROUP_SORT_COLUMN,
  );

// Deleting a bidirectional link also deletes its reverse link on the target
// cluster; deleting a unidirectional link also deletes its link identity there.
export const hasTargetCleanup = (link: ClusterLink): boolean =>
  isBidirectional(link) || isUnidirectional(link);

export const invalidateClusterLinkLists = async (
  queryClient: QueryClient,
  clusters: Iterable<string>,
): Promise<void> => {
  const affectedClusters = new Set(clusters);
  if (affectedClusters.size === 0) {
    return;
  }

  await Promise.all([
    queryClient.invalidateQueries({ queryKey: allClusterLinksQueryKey }),
    ...Array.from(affectedClusters, async (cluster) =>
      queryClient.invalidateQueries({
        queryKey: [queryKeys.clusters, cluster, queryKeys.links],
      }),
    ),
  ]);
};

export const invalidateClusterLinksAndIdentities = async (
  queryClient: QueryClient,
  sourceCluster: string,
  link: ClusterLink,
  affectedTargetCluster?: string,
): Promise<void> => {
  const linkClusters = new Set([sourceCluster]);
  const identityClusters = new Set<string>();

  const bidirectional = isBidirectional(link);

  if (bidirectional) {
    identityClusters.add(sourceCluster);
  }
  if (affectedTargetCluster !== undefined && hasTargetCleanup(link)) {
    identityClusters.add(affectedTargetCluster);
    if (bidirectional) {
      linkClusters.add(affectedTargetCluster);
    }
  }

  await Promise.all([
    invalidateClusterLinkLists(queryClient, linkClusters),
    ...Array.from(identityClusters, async (cluster) =>
      queryClient.invalidateQueries({
        queryKey: [queryKeys.clusters, cluster, queryKeys.identities],
      }),
    ),
  ]);
};

// The target cluster can only be cleaned up through its registered tunnel.
export const canCleanUpTarget = (
  link: ClusterLink,
  targetCluster?: Cluster,
): boolean => hasTargetCleanup(link) && !!targetCluster?.tunnel_registered;

export const findCluster = (
  clusters: Cluster[],
  name: string | null,
): Cluster | undefined =>
  name ? clusters.find((cluster) => cluster.name === name) : undefined;

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

// Public links never present a client certificate, so LXD reports a reachable remote as "Unauthenticated"
const isReachableMember = (
  member: ClusterLinkMemberState,
  link?: ClusterLink,
): boolean => {
  return (
    member.status === "Active" ||
    (link?.type === "public" && member.status === "Unauthenticated")
  );
};

// A public link stays pending until its certificate is confirmed, which sets volatile.addresses
export const isPendingPublicLink = (link?: ClusterLink): boolean => {
  return link?.type === "public" && !link?.config?.["volatile.addresses"];
};

export const getClusterLinksStatus = (
  identity?: LxdIdentity,
  state?: ClusterLinkState,
  link?: ClusterLink,
): StatusCaption => {
  if (identity?.type.toLowerCase().includes("(pending)")) {
    return "Pending";
  }
  if (link && isPendingPublicLink(link)) {
    return "Pending";
  }
  if (
    state?.cluster_link_members.some((member) =>
      isReachableMember(member, link),
    )
  ) {
    return "Reachable";
  }
  return "Unreachable";
};

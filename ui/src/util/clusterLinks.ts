import type { QueryClient } from "@tanstack/react-query";
import type {
  Cluster,
  ClusterLink,
  ClusterLinkMemberState,
  ClusterLinkState,
  ResolvedClusterLink,
  StatusCaption,
} from "types/cluster";
import { queryKeys } from "util/queryKeys";
import type { LxdIdentity } from "types/permissions";

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

export const getClusterLinkId = ({
  source_cluster,
  link,
}: ResolvedClusterLink): string => `${source_cluster}/${link.name}`;

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

// Deleting a bidirectional link also deletes its reverse link on the target
// cluster; deleting a unidirectional link also deletes its link identity there.
export const hasTargetCleanup = (link: ClusterLink): boolean =>
  isBidirectional(link) || isUnidirectional(link);

export const invalidateClusterLinkLists = async (
  queryClient: QueryClient,
  clusters: Iterable<string>,
): Promise<void> => {
  await Promise.all(
    Array.from(new Set(clusters), async (cluster) =>
      queryClient.invalidateQueries({
        queryKey: [queryKeys.clusters, cluster, queryKeys.links],
      }),
    ),
  );
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

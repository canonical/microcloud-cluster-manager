import type { QueryClient } from "@tanstack/react-query";
import type {
  Cluster,
  ClusterLink,
  ClusterLinkMemberState,
  ClusterLinkState,
  StatusCaption,
} from "types/cluster";
import type { LxdIdentity } from "types/permissions";
import { queryKeys } from "util/queryKeys";

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

// A connection over a cluster link: the link itself, the source cluster it is
// defined on, and the target cluster it points to, if the target is known to
// the cluster manager.
export interface ClusterConnection {
  id: string;
  source: Cluster;
  link: ClusterLink;
  target?: Cluster;
}

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
  addresses: getLinkDisplayAddresses(entry.link).join(",").toLowerCase(),
});

const LXD_DEFAULT_PORT = 8443;

// Converts a registered LXD URL to the host and port stored in a cluster link.
const getClusterAddress = (lxdUrl: string): string | undefined => {
  try {
    const url = lxdUrl.trim();
    const { hostname, port } = new URL(url);
    // URL.port drops an explicit scheme-default port such as HTTPS port 443.
    const explicitPort = /^https?:\/\/[^/?#]*:(\d+)(?:[/?#]|$)/i.exec(url)?.[1];
    const effectivePort = Number(port || explicitPort || LXD_DEFAULT_PORT);
    return hostname ? `${hostname.toLowerCase()}:${effectivePort}` : undefined;
  } catch {
    return undefined;
  }
};

const hasLinkAddress = (lxdUrl: string, addresses: string[]): boolean => {
  const clusterAddress = getClusterAddress(lxdUrl);
  return (
    !!clusterAddress &&
    addresses.some((address) => address.toLowerCase() === clusterAddress)
  );
};

// Resolves the cluster a link points at by UUID, falling back to matching its
// confirmed addresses against the LXD URLs of known clusters.
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

  const addresses = getLinkAddresses(link);
  return clusters.find((cluster) => hasLinkAddress(cluster.lxd_url, addresses));
};

// Finds the bidirectional link on the other end of a cluster link, meaning the
// link on the linked cluster that points back at the given cluster. Other link
// types have no counterpart on the linked cluster.
export const getReverseLink = (
  links: ClusterLink[],
  cluster: Cluster,
): ClusterLink | undefined => {
  return links.find((link) => {
    if (!isBidirectional(link)) {
      return false;
    }

    return (
      link.config?.["volatile.uuid"] === cluster.cluster_uuid ||
      hasLinkAddress(cluster.lxd_url, getLinkAddresses(link))
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
  if (affectedTargetCluster !== undefined && hasTargetEnd(link)) {
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

export const findClusterLinkIdentity = (
  identities: LxdIdentity[],
  fingerprint: string,
): LxdIdentity | undefined => {
  return identities.find(
    (identity) =>
      identity.authentication_method === "tls" &&
      identity.type === "Cluster link certificate" &&
      identity.id.toLowerCase() === fingerprint.toLowerCase(),
  );
};

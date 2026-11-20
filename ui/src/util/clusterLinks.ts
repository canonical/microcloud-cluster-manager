import type { LxdIdentity } from "types/permissions";
import type { Cluster, ClusterLink, ResolvedClusterLink } from "types/cluster";

export const getLinkAddresses = (link: ClusterLink): string[] => {
  const addresses =
    link.config?.["volatile.addresses"]
      ?.split(",")
      .map((address) => address.trim())
      .filter(Boolean) ?? [];
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
            ...getLinkAddresses(link),
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

export const getClusterLinkSortData = ({
  link,
  target_cluster,
}: ResolvedClusterLink) => ({
  name: link.name.toLowerCase(),
  type: link.type.toLowerCase(),
  target: (target_cluster ?? "").toLowerCase(),
  description: (link.description ?? "").toLowerCase(),
  usedBy: link.used_by?.length ?? 0,
  addresses: getLinkAddresses(link).join(",").toLowerCase(),
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

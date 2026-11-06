import type {
  Cluster,
  ClusterLink,
  ClusterLinkMemberState,
  ClusterLinkState,
  StatusCaption,
} from "types/cluster";
import type { LxdIdentity } from "types/permissions";

const getLinkAddresses = (link: ClusterLink): string[] => {
  return link.config?.["volatile.addresses"]?.split(",") ?? [];
};

export const isPendingPublicLink = (link: ClusterLink): boolean => {
  return link.type === "public" && !link.config?.["volatile.addresses"];
};

export const getCertificateFingerprint = async (
  pem: string,
): Promise<string> => {
  const base64 = pem
    .replace(/-----BEGIN CERTIFICATE-----|-----END CERTIFICATE-----/g, "")
    .replace(/\s/g, "");
  const der = Uint8Array.from(atob(base64), (character) =>
    character.charCodeAt(0),
  );
  const digest = await crypto.subtle.digest("SHA-256", der);

  return [...new Uint8Array(digest)]
    .map((byte) => byte.toString(16).padStart(2, "0"))
    .join("");
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

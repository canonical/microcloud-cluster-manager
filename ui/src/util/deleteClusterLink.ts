import { deleteClusterLink, fetchClusterLinks } from "api/clusters";
import type { Cluster, ClusterLink } from "types/cluster";
import { getReverseLink } from "util/clusterLinks";

// Deletes a link and, when requested, the link pointing back on the target cluster.
export const deleteClusterLinkWithTarget = async (
  cluster: Cluster,
  link: ClusterLink,
  targetCluster?: Cluster,
  removeTargetEnd = false,
): Promise<void> => {
  await deleteClusterLink(cluster.name, link.name);

  if (!targetCluster || !removeTargetEnd) {
    return;
  }

  const targetLinks = await fetchClusterLinks(targetCluster.name);
  const reverseLink = getReverseLink(targetLinks, cluster);
  if (!reverseLink) {
    throw new Error(
      `No cluster link pointing back at ${cluster.name} found on cluster ${targetCluster.name}.`,
    );
  }

  await deleteClusterLink(targetCluster.name, reverseLink.name);
};

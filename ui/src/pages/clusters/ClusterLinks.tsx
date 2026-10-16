import { Notification } from "@canonical/react-components";
import { useQuery } from "@tanstack/react-query";
import { fetchClusterLinks, fetchClusters } from "api/clusters";
import type { FC } from "react";
import { queryKeys } from "util/queryKeys";
import type { Cluster } from "types/cluster";
import ClusterLinkCard from "pages/clusters/ClusterLinkCard";
import CreateClusterLinkBtn from "pages/clusters/actions/CreateClusterLinkBtn";
import ClusterLinksEmptyState from "components/ClusterLinksEmptyState";

interface Props {
  cluster: Cluster;
}

const ClusterLinks: FC<Props> = ({ cluster }) => {
  const {
    data: clusterLinks,
    error,
    isLoading,
  } = useQuery({
    queryKey: [queryKeys.clusters, cluster.name, queryKeys.links],
    queryFn: async () => fetchClusterLinks(cluster.name),
    enabled: cluster.tunnel_registered,
  });

  const { data: clusters = [], isLoading: isClustersLoading } = useQuery({
    queryKey: [queryKeys.clusters],
    queryFn: fetchClusters,
  });

  if (isLoading || isClustersLoading) {
    return <></>;
  }

  return (
    <>
      {!cluster.tunnel_registered && (
        <Notification
          severity="caution"
          title="Cluster links require a registered tunnel."
        >
          Run <code>microcloud cluster-manager set reverse_tunnel true</code> on
          this cluster to enable the tunnel.
        </Notification>
      )}
      {cluster.tunnel_registered && error && (
        <Notification severity="negative" title="Loading cluster links failed">
          {error.message}
        </Notification>
      )}
      {cluster.tunnel_registered && clusterLinks && clusterLinks.length > 0 && (
        <>
          <div className="cluster-links-header">
            <h2 className="p-heading--4 u-no-margin--bottom u-no-padding--top">
              Cluster links
            </h2>
            <CreateClusterLinkBtn cluster={cluster} />
          </div>
          {clusterLinks.map((link) => (
            <ClusterLinkCard
              key={link.name}
              cluster={cluster}
              link={link}
              clusters={clusters}
            />
          ))}
        </>
      )}
      {cluster.tunnel_registered &&
        !error &&
        clusterLinks &&
        clusterLinks.length === 0 && (
          <ClusterLinksEmptyState cluster={cluster} />
        )}
    </>
  );
};

export default ClusterLinks;

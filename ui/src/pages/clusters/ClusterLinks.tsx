import { Notification } from "@canonical/react-components";
import { useQuery } from "@tanstack/react-query";
import { fetchClusters } from "api/clusters";
import type { FC } from "react";
import { useClusterLinks } from "context/useClusters";
import { queryKeys } from "util/queryKeys";
import type { Cluster } from "types/cluster";
import ClusterLinksTable from "pages/clusters/ClusterLinksTable";
import ClusterLinksEmptyState from "components/ClusterLinksEmptyState";

interface Props {
  cluster: Cluster;
}

const ClusterLinks: FC<Props> = ({ cluster }) => {
  const {
    data: clusterLinks,
    error,
    isLoading,
  } = useClusterLinks(cluster.name, cluster.tunnel_registered);

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
        <ClusterLinksTable
          cluster={cluster}
          clusters={clusters}
          links={clusterLinks}
        />
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

import { Notification } from "@canonical/react-components";
import type { FC } from "react";
import { useClusterLinks } from "context/useClusters";
import type { Cluster } from "types/cluster";
import ClusterLinksTable from "pages/clusters/ClusterLinksTable";
import ClusterLinksEmptyState from "pages/clusters/ClusterLinksEmptyState";

interface Props {
  cluster: Cluster;
}

const ClusterLinksTab: FC<Props> = ({ cluster }) => {
  const {
    data: clusterLinks,
    error,
    isLoading,
  } = useClusterLinks(cluster.name, cluster.tunnel_registered);

  if (isLoading) {
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
          key={cluster.name}
          cluster={cluster}
          clusterLinks={clusterLinks}
        />
      )}
      {cluster.tunnel_registered &&
        !error &&
        clusterLinks &&
        clusterLinks.length === 0 && (
          <ClusterLinksEmptyState cluster={cluster.name} />
        )}
    </>
  );
};

export default ClusterLinksTab;

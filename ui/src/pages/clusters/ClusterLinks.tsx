import { Notification } from "@canonical/react-components";
import { useQuery } from "@tanstack/react-query";
import { fetchClusterLinks, fetchClusters } from "api/clusters";
import type { FC } from "react";
import { queryKeys } from "util/queryKeys";
import type { Cluster } from "types/cluster";
import ClusterLinkCard from "pages/clusters/ClusterLinkCard";
import CreateClusterLinkBtn from "pages/clusters/actions/CreateClusterLinkBtn";

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
      <h2 className="p-heading--4">Cluster Links</h2>
      {!cluster.tunnel_registered && (
        <Notification
          severity="caution"
          title="Not available without registered tunnel."
        />
      )}
      {cluster.tunnel_registered && error && (
        <Notification severity="negative" title="Loading cluster links failed">
          {error.message}
        </Notification>
      )}
      {cluster.tunnel_registered && clusterLinks && clusterLinks.length > 0 ? (
        clusterLinks.map((link) => (
          <ClusterLinkCard
            key={link.name}
            cluster={cluster}
            link={link}
            clusters={clusters}
          />
        ))
      ) : (
        <Notification severity="information" title="No cluster links found." />
      )}
      {cluster.tunnel_registered && <CreateClusterLinkBtn cluster={cluster} />}
    </>
  );
};

export default ClusterLinks;

import type { FC } from "react";
import { Notification, Spinner } from "@canonical/react-components";
import { useQueries, useQuery } from "@tanstack/react-query";
import { fetchClusterLinks, fetchClusters } from "api/clusters";
import classNames from "classnames";
import ResourceLink from "components/ResourceLink";
import type { Cluster, ClusterLink } from "types/cluster";
import { queryKeys } from "util/queryKeys";
import {
  hasOnwardClusterLinks,
  toClusterLinkTopologyItem,
} from "util/clusterLinkTopology";
import ClusterLinkRichChip from "./ClusterLinkRichChip";
import ClusterRichChip from "./ClusterRichChip";

export interface Props {
  cluster: Cluster;
}

export const ClusterTopology: FC<Props> = ({ cluster }) => {
  const {
    data: clusters = [],
    error: clustersError,
    isLoading: isClustersLoading,
  } = useQuery({
    queryKey: [queryKeys.clusters],
    queryFn: fetchClusters,
  });

  const {
    data: topologyItems = [],
    error,
    isLoading,
  } = useQuery({
    queryKey: [queryKeys.clusters, cluster.name, queryKeys.links],
    queryFn: async () => fetchClusterLinks(cluster.name),
    enabled: cluster.tunnel_registered,
    select: (clusterLinks) =>
      clusterLinks.map((link) => toClusterLinkTopologyItem(link, clusters)),
  });

  const targetNames = [
    ...new Set(
      topologyItems.flatMap((link) =>
        link.targetClusterName ? [link.targetClusterName] : [],
      ),
    ),
  ];
  const targetQueries = useQueries({
    queries: targetNames.map((name) => ({
      queryKey: [queryKeys.clusters, name, queryKeys.links],
      queryFn: async () => fetchClusterLinks(name),
      select: (targetLinks: ClusterLink[]) =>
        hasOnwardClusterLinks(targetLinks, cluster),
    })),
  });
  const targetError = targetQueries.find((query) => query.error)?.error;
  const links = topologyItems.map((link) => ({
    ...link,
    hasMore: link.targetClusterName
      ? targetQueries[targetNames.indexOf(link.targetClusterName)]?.data
      : link.hasMore,
  }));

  if (error || clustersError || targetError) {
    return (
      <Notification severity="negative" title="Loading cluster topology failed">
        {(error ?? clustersError ?? targetError)?.message}
      </Notification>
    );
  }

  if (
    isLoading ||
    isClustersLoading ||
    targetQueries.some((query) => query.isLoading)
  ) {
    return <Spinner text="Loading cluster topology..." />;
  }

  return (
    <div className="cluster-topology">
      <div
        className={classNames("current-cluster", {
          "has-descendents": links.length > 0,
        })}
      >
        <ResourceLink
          type="cluster"
          value={cluster.name}
          to="#"
          className="active-chip"
        />
      </div>

      {links.length > 0 && (
        <div className="downstream">
          {links.map((link) => (
            <div key={link.id} className="downstream-item">
              <div className="link-cell">
                <ClusterLinkRichChip
                  cluster={cluster.name}
                  clusterLink={link.clusterLink}
                />
                <span className="target-connector" />
              </div>

              <div className="target-cell">
                {link.targetClusterName ? (
                  <>
                    <ClusterRichChip
                      clusterName={link.targetClusterName}
                      urlSuffix="/cluster-links"
                    />
                    {link.hasMore && (
                      <>
                        <span className="target-connector" />
                        <span className="trail-dots">⋯</span>
                      </>
                    )}
                  </>
                ) : (
                  link.hasMore && <span className="trail-dots">⋯</span>
                )}
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
};

export default ClusterTopology;

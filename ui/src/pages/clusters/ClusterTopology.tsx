import { useState, type FC } from "react";
import { Button, Notification, Spinner } from "@canonical/react-components";
import { useQuery } from "@tanstack/react-query";
import classnames from "classnames";
import { fetchClusterLinks, fetchClusters } from "api/clusters";
import CreateClusterLinkBtn from "components/cluster-links/CreateClusterLinkBtn";
import ClusterLinkRichChip from "pages/clusters/ClusterLinkRichChip";
import ClusterRichChip from "pages/clusters/ClusterRichChip";
import type { Cluster } from "types/cluster";
import { queryKeys } from "util/queryKeys";
import { toClusterLinkTopologyItem } from "util/clusterTopology";

export interface Props {
  cluster: Cluster;
}

const MAX_COLLAPSED_LINKS = 4;

const ClusterTopology: FC<Props> = ({ cluster }) => {
  const [isExpanded, setIsExpanded] = useState(false);

  const {
    data: clusters = [],
    error: clustersError,
    isLoading: isClustersLoading,
  } = useQuery({
    queryKey: [queryKeys.clusters],
    queryFn: fetchClusters,
  });

  const {
    data: clusterLinksRaw = [],
    error,
    isLoading: isRawLinksLoading,
  } = useQuery({
    queryKey: [queryKeys.clusters, cluster.name, queryKeys.links, "topology"],
    queryFn: async () => fetchClusterLinks(cluster.name, "2"),
    enabled: Boolean(cluster.tunnel_registered),
  });

  const enrichedLinks = clusterLinksRaw.map((link) =>
    toClusterLinkTopologyItem(link, clusters),
  );

  const shouldTruncate =
    enrichedLinks.length > MAX_COLLAPSED_LINKS && !isExpanded;
  const displayedLinks = shouldTruncate
    ? enrichedLinks.slice(0, MAX_COLLAPSED_LINKS)
    : enrichedLinks;

  if (error || clustersError) {
    return (
      <Notification severity="negative" title="Loading cluster topology failed">
        {(error ?? clustersError)?.message}
      </Notification>
    );
  }

  const isLinksLoading =
    Boolean(cluster.tunnel_registered) && isRawLinksLoading;

  if (isClustersLoading || isLinksLoading) {
    return <Spinner text="Loading cluster topology..." />;
  }

  return (
    <>
      <div className="cluster-topology-header">
        <h2
          className={classnames("p-heading--4 u-no-padding--top", {
            "u-no-margin--bottom": clusterLinksRaw.length !== 0,
          })}
        >
          Topology
        </h2>
        <CreateClusterLinkBtn cluster={cluster.name} />
      </div>
      <div className="cluster-topology">
        <div
          className={classnames("current-cluster", {
            "has-descendants": enrichedLinks.length > 0,
          })}
        >
          <ClusterRichChip
            clusterName={cluster.name}
            className="active-chip"
            urlSuffix="/cluster-links"
          />
        </div>

        {enrichedLinks.length > 0 && (
          <div className="downstream">
            {displayedLinks.map((link) => (
              <div key={link.id} className="downstream-item">
                <div className="link-cell">
                  <ClusterLinkRichChip
                    cluster={cluster.name}
                    clusterLinkName={link.clusterLink.name}
                    clusterLinkType={link.clusterLink.type}
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

            {shouldTruncate && (
              <Button
                appearance="base"
                className="show-more-button"
                onClick={() => {
                  setIsExpanded(true);
                }}
              >
                Show {enrichedLinks.length - MAX_COLLAPSED_LINKS} more
              </Button>
            )}

            {isExpanded && enrichedLinks.length > MAX_COLLAPSED_LINKS && (
              <Button
                appearance="base"
                className="hide-button"
                onClick={() => {
                  setIsExpanded(false);
                }}
              >
                Hide {enrichedLinks.length - MAX_COLLAPSED_LINKS}
              </Button>
            )}
          </div>
        )}
      </div>
    </>
  );
};

export default ClusterTopology;

import { type FC } from "react";
import { Spinner } from "@canonical/react-components";
import { useQuery } from "@tanstack/react-query";
import { Link } from "react-router-dom";
import { fetchClusters } from "api/clusters";
import { useClusterLink } from "context/useClusterLinks";
import type { TooltipRow } from "components/RichTooltipRow";
import { RichTooltipTable } from "components/RichTooltipTable";
import ClusterLinkStatus from "./ClusterLinkStatus";
import { getLinkedCluster } from "util/clusterLinks";
import { queryKeys } from "util/queryKeys";

interface Props {
  cluster: string;
  clusterLinkName: string;
}

const ClusterLinkRichTooltip: FC<Props> = ({ cluster, clusterLinkName }) => {
  const { data: clusterLink, isLoading: isLinkLoading } = useClusterLink(
    cluster,
    clusterLinkName,
  );
  const {
    data: clusters = [],
    isLoading: isClustersLoading,
    error: clustersError,
  } = useQuery({
    queryKey: [queryKeys.clusters],
    queryFn: fetchClusters,
  });

  if (!clusterLink && !isLinkLoading) {
    return <>Cluster link {clusterLinkName} not found</>;
  }

  const targetCluster = clusterLink
    ? getLinkedCluster(clusterLink, clusters)
    : undefined;

  const rows: TooltipRow[] = [
    {
      title: "Cluster link",
      value: clusterLink ? (
        <Link to={`/ui/cluster/${encodeURIComponent(cluster)}/cluster-links`}>
          {clusterLinkName}
        </Link>
      ) : (
        <Spinner />
      ),
      valueTitle: clusterLinkName,
    },
    {
      title: "Status",
      value: clusterLink ? (
        <ClusterLinkStatus cluster={cluster} link={clusterLink} />
      ) : (
        "-"
      ),
      className: "status-row",
    },
    {
      title: "Description",
      value: clusterLink?.description || "-",
      valueTitle: clusterLink?.description || "",
    },
    {
      title: "Source",
      value: (
        <Link to={`/ui/cluster/${encodeURIComponent(cluster)}`}>{cluster}</Link>
      ),
    },
    {
      title: "Target",
      value:
        isLinkLoading || isClustersLoading ? (
          <Spinner />
        ) : clustersError ? (
          <span title={clustersError.message}>
            Loading target cluster failed
          </span>
        ) : targetCluster ? (
          <Link to={`/ui/cluster/${encodeURIComponent(targetCluster.name)}`}>
            {targetCluster.name}
          </Link>
        ) : (
          "-"
        ),
    },
  ];

  return (
    <RichTooltipTable rows={rows} className="cluster-link-rich-tooltip-table" />
  );
};

export default ClusterLinkRichTooltip;

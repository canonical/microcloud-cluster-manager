import { type FC } from "react";
import { Spinner } from "@canonical/react-components";
import { Link } from "react-router-dom";
import { useClusterLinks } from "context/useClusterLinks";
import type { TooltipRow } from "components/RichTooltipRow";
import { RichTooltipTable } from "components/RichTooltipTable";
import ClusterLinkStatus from "./ClusterLinkStatus";

interface Props {
  cluster: string;
  clusterLinkName: string;
}

const ClusterLinkRichTooltip: FC<Props> = ({ cluster, clusterLinkName }) => {
  const {
    data: links = [],
    isLoading: isLinkLoading,
    error,
  } = useClusterLinks(cluster);
  const resolvedLink = links.find(({ link }) => link.name === clusterLinkName);
  const clusterLink = resolvedLink?.link;
  const targetCluster = resolvedLink?.target_cluster;

  if (error) {
    return <span title={error.message}>Loading cluster link failed</span>;
  }

  if (!clusterLink && !isLinkLoading) {
    return <>Cluster link {clusterLinkName} not found</>;
  }

  let targetValue: TooltipRow["value"] = "-";

  if (isLinkLoading) {
    targetValue = <Spinner />;
  } else if (targetCluster) {
    targetValue = (
      <Link to={`/ui/cluster/${encodeURIComponent(targetCluster)}`}>
        {targetCluster}
      </Link>
    );
  }

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
      value: targetValue,
    },
  ];

  return (
    <RichTooltipTable rows={rows} className="cluster-link-rich-tooltip-table" />
  );
};

export default ClusterLinkRichTooltip;

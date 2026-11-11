import { type FC } from "react";
import { Spinner } from "@canonical/react-components";
import { useQuery } from "@tanstack/react-query";
import { Link } from "react-router-dom";
import { fetchClusters } from "api/clusters";
import { useClusterIdentities } from "context/useClusterIdentities";
import { useClusterLink } from "context/useClusterLinks";
import type { TooltipRow } from "components/RichTooltipRow";
import { RichTooltipTable } from "components/RichTooltipTable";
import ItemName from "components/ItemName";
import ClusterLinkStatus from "./ClusterLinkStatus";
import ResourceLabel from "components/ResourceLabel";
import ClusterLinkAddresses from "./ClusterLinkAddresses";
import { getLinkedCluster, getLinkIdentity } from "util/clusterLinks";
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
  const { data: identities = [] } = useClusterIdentities(cluster);
  const {
    data: clusters = [],
    isLoading: isClustersLoading,
    error: clustersError,
  } = useQuery({
    queryKey: [queryKeys.clusters],
    queryFn: fetchClusters,
  });

  if (!clusterLink && !isLinkLoading) {
    return (
      <>
        Cluster link{" "}
        <ResourceLabel type="cluster-link-public" value={clusterLinkName} /> not
        found
      </>
    );
  }

  const identity = getLinkIdentity(identities, clusterLinkName);
  const authGroupCount = identity?.groups?.length ?? 0;
  const targetCluster = clusterLink
    ? getLinkedCluster(clusterLink, clusters)
    : undefined;

  const rows: TooltipRow[] = [
    {
      title: "Cluster link",
      value: clusterLink ? (
        <ItemName item={{ name: clusterLinkName }} />
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
      title: "Type",
      value: clusterLink?.type || "-",
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
    {
      title: "Addresses",
      value: clusterLink ? (
        <ClusterLinkAddresses clusterLink={clusterLink} />
      ) : (
        "-"
      ),
    },
    {
      title: "Auth groups",
      value: authGroupCount,
    },
  ];

  return (
    <RichTooltipTable rows={rows} className="cluster-link-rich-tooltip-table" />
  );
};

export default ClusterLinkRichTooltip;

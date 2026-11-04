import { type FC } from "react";
import { Spinner } from "@canonical/react-components";
import type { TooltipRow } from "components/RichTooltipRow";
import { RichTooltipTable } from "components/RichTooltipTable";
import ItemName from "components/ItemName";
import { useClusterLink } from "context/useClusters";
import ClusterLinkStatus from "./ClusterLinkStatus";
import ResourceLabel from "components/ResourceLabel";
import ClusterLinkAddresses from "./ClusterLinkAddresses";

interface Props {
  cluster: string;
  clusterLinkName: string;
}

const ClusterLinkRichTooltip: FC<Props> = ({ cluster, clusterLinkName }) => {
  const { data: clusterLink, isLoading: isLinkLoading } = useClusterLink(
    cluster,
    clusterLinkName,
  );

  if (!clusterLink && !isLinkLoading) {
    return (
      <>
        Cluster link{" "}
        <ResourceLabel type="cluster-link-public" value={clusterLinkName} /> not
        found
      </>
    );
  }

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
      title: "Addresses",
      value: clusterLink ? (
        <ClusterLinkAddresses clusterLink={clusterLink} />
      ) : (
        "-"
      ),
    },
    {
      title: "Auth groups",
      value: clusterLink ? (clusterLink.auth_groups_count ?? 0) : <Spinner />,
    },
  ];

  return (
    <RichTooltipTable rows={rows} className="cluster-link-rich-tooltip-table" />
  );
};

export default ClusterLinkRichTooltip;

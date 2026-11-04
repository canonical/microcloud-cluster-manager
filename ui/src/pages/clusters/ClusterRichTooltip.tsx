import type { FC } from "react";
import { Spinner } from "@canonical/react-components";
import type { TooltipRow } from "components/RichTooltipRow";
import { RichTooltipTable } from "components/RichTooltipTable";
import ItemName from "components/ItemName";
import { useCluster } from "context/useClusters";
import ClusterStatus from "./ClusterStatus";

interface Props {
  clusterName: string;
}

const ClusterRichTooltip: FC<Props> = ({ clusterName }) => {
  const { data: cluster, isLoading } = useCluster(clusterName);

  if (!cluster && !isLoading) {
    return <>Cluster {clusterName} not found</>;
  }

  const runningInstances =
    cluster?.instance_statuses?.find((s) => s.status === "Running")?.count ?? 0;

  const rows: TooltipRow[] = [
    {
      title: "Cluster",
      value: cluster ? <ItemName item={{ name: cluster.name }} /> : <Spinner />,
      valueTitle: cluster?.name || "",
    },
    {
      title: "Status",
      value: cluster ? <ClusterStatus cluster={cluster} /> : "-",
      className: "status-row",
    },
    {
      title: "Description",
      value: cluster?.description || "-",
      valueTitle: cluster?.description || "",
    },
    {
      title: "Members",
      value: cluster?.member_count ?? 0,
    },
    {
      title: "Running instances",
      value: runningInstances,
    },
  ];

  return (
    <RichTooltipTable rows={rows} className="cluster-rich-tooltip-table" />
  );
};

export default ClusterRichTooltip;

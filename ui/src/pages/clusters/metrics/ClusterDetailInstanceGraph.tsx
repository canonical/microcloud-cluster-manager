import { Icon } from "@canonical/react-components";
import DoughnutChart from "components/DoughnutChart";
import { FC, ReactNode } from "react";
import { Cluster } from "types/cluster";
import { statusCount } from "util/helpers";

interface Props {
  cluster: Cluster;
}

const ClusterDetailInstanceGraph: FC<Props> = ({ cluster }: Props) => {
  const runningInstances = statusCount(cluster.instance_statuses, "Running");
  const stoppedInstances = statusCount(cluster.instance_statuses, "Stopped");
  const frozenInstances = statusCount(cluster.instance_statuses, "Frozen");
  const errorInstances = statusCount(cluster.instance_statuses, "Error");

  const totalInstances =
    runningInstances + stoppedInstances + frozenInstances + errorInstances;

  function getPercentageString(portion: number): ReactNode {
    return (
      <>
        <b>{portion}</b>{" "}
        {totalInstances > 0
          ? `(${Math.floor((portion / totalInstances) * 100)}%) `
          : ""}
      </>
    );
  }

  return (
    <div className="cluster-detail-doughnut-graph">
      <DoughnutChart
        chartID="clusterInstance"
        segmentHoverWidth={45}
        segmentWidth={40}
        segments={[
          {
            color: "#0E8420",
            tooltip: "Running",
            value: runningInstances,
          },
          {
            color: "#CC7900",
            tooltip: "Stopped",
            value: stoppedInstances,
          },
          { color: "#24598f", tooltip: "Frozen", value: frozenInstances },
          { color: "#C7162B", tooltip: "Error", value: errorInstances },
        ]}
        size={150}
      />
      <ul className="doughnut-chart__legend u-no-margin--left">
        <li className="u-no-margin p-heading--5 u-no-padding">
          {totalInstances} Instances
        </li>
        <li>
          <Icon name="status-succeeded-small" />
          {getPercentageString(runningInstances)} Running
        </li>
        <li>
          <Icon name="status-waiting-small" />
          {getPercentageString(stoppedInstances)}
          Stopped
        </li>
        <li>
          <Icon name="status-in-progress-small" />
          {getPercentageString(frozenInstances)}
          Frozen
        </li>
        <li>
          <Icon name="status-failed-small" />
          {getPercentageString(errorInstances)}
          Error
        </li>
      </ul>
    </div>
  );
};

export default ClusterDetailInstanceGraph;

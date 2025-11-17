import { DoughnutChart, Icon } from "@canonical/react-components";
import type { FC, ReactNode } from "react";
import type { Cluster } from "types/cluster";
import { pluralize, statusCount } from "util/helpers";

interface Props {
  cluster: Cluster;
}

const ClusterDetailInstanceGraph: FC<Props> = ({ cluster }: Props) => {
  const running = statusCount(cluster.instance_statuses, "Running");
  const stopped = statusCount(cluster.instance_statuses, "Stopped");
  const frozen = statusCount(cluster.instance_statuses, "Frozen");
  const error = statusCount(cluster.instance_statuses, "Error");

  const total = running + stopped + frozen + error;

  const getPercentage = (portion: number): ReactNode => (
    <>
      <b>{portion}</b>{" "}
      {total > 0 ? `(${Math.floor((portion / total) * 100)}%) ` : ""}
    </>
  );

  const segments = [{ color: "#D3E4ED", tooltip: "", value: 0 }];
  if (running > 0) {
    segments.push({
      color: "#0E8420",
      tooltip: "Running",
      value: running,
    });
  }
  if (stopped > 0) {
    segments.push({
      color: "#CC7900",
      tooltip: "Stopped",
      value: stopped,
    });
  }
  if (frozen > 0) {
    segments.push({
      color: "#24598f",
      tooltip: "Frozen",
      value: frozen,
    });
  }
  if (error > 0) {
    segments.push({
      color: "#C7162B",
      tooltip: "Error",
      value: error,
    });
  }

  return (
    <div className="cluster-detail-doughnut-graph">
      <DoughnutChart
        chartID="clusterInstance"
        segmentHoverWidth={45}
        segmentThickness={40}
        segments={segments}
        size={150}
      />
      <ul className="doughnut-chart__legend u-no-margin--left">
        <li className="u-no-margin p-heading--5 u-no-padding">
          {total} {pluralize("instance", total)}
        </li>
        <li>
          <Icon name="status-succeeded-small" />
          {getPercentage(running)} Running
        </li>
        <li>
          <Icon name="status-waiting-small" />
          {getPercentage(stopped)}
          Stopped
        </li>
        <li>
          <Icon name="status-in-progress-small" />
          {getPercentage(frozen)}
          Frozen
        </li>
        <li>
          <Icon name="status-failed-small" />
          {getPercentage(error)}
          Error
        </li>
      </ul>
    </div>
  );
};

export default ClusterDetailInstanceGraph;

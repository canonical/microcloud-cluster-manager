import { Icon } from "@canonical/react-components";
import DoughnutChart from "components/DoughnutChart";
import { FC, ReactNode } from "react";
import { Cluster } from "types/cluster";
import { getMinutesSinceLastHeartbeat } from "util/helpers";

interface Props {
  clusters: Cluster[];
}

const ClusterStatusGraph: FC<Props> = ({ clusters }) => {
  const totalClusters = clusters.length;
  const activeClusters = clusters.filter(
    (cluster) =>
      cluster.status == "ACTIVE" && getMinutesSinceLastHeartbeat(cluster) < 5,
  ).length;
  const pendingClusters = clusters.filter(
    (cluster) => cluster.status == "PENDING_APPROVAL",
  ).length;
  const degradedClusters = clusters.filter(
    (cluster) =>
      cluster.status == "ACTIVE" && getMinutesSinceLastHeartbeat(cluster) > 5,
  ).length;

  function getPercentageString(portion: number): ReactNode {
    return (
      <>
        <b>{portion}</b> {`(${Math.floor((portion / totalClusters) * 100)}%) `}
      </>
    );
  }

  return (
    <div className="cluster-doughnut-graph">
      <DoughnutChart
        chartID="clusterStatus"
        segmentHoverWidth={45}
        segmentWidth={40}
        segments={[
          { color: "#0E8420", tooltip: "Active", value: activeClusters },
          { color: "#CC7900", tooltip: "Pending", value: pendingClusters },
          { color: "#C7162B", tooltip: "Degraded", value: degradedClusters },
        ]}
        size={150}
      />
      <ul className="doughnut-chart__legend u-no-margin--left">
        <li className="u-no-margin p-heading--5 u-no-padding">
          {totalClusters} clusters
        </li>
        <li>
          <Icon name="status-succeeded-small" />
          {getPercentageString(activeClusters)} Online
        </li>
        <li>
          <Icon name="status-waiting-small" />
          {getPercentageString(pendingClusters)}
          Pending
        </li>
        <li>
          <Icon name="status-failed-small" />
          {getPercentageString(degradedClusters)}
          Degraded
        </li>
      </ul>
    </div>
  );
};

export default ClusterStatusGraph;

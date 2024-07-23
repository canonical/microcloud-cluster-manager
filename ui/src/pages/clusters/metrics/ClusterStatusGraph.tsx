import { Icon } from "@canonical/react-components";
import { useQuery } from "@tanstack/react-query";
import { fetchClusters } from "api/clusters";
import DoughnutChart from "components/DoughnutChart";
import { FC, ReactNode } from "react";
import { getMinutesSinceLastHeartbeat } from "util/helpers";
import { queryKeys } from "util/queryKeys";

const ClusterStatusGraph: FC = () => {
  const { data: clusters = [] } = useQuery({
    queryKey: [queryKeys.clusters],
    queryFn: fetchClusters,
  });

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
    <div className="cluster-status">
      <DoughnutChart
        className="cluster-status__chart"
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
      <div className="cluster-status-legend">
        <div className="u-no-margin p-heading--5 u-no-padding">
          {totalClusters} clusters
        </div>
        <div>
          <Icon name="status-succeeded-small" />
          {getPercentageString(activeClusters)} Online
        </div>
        <div>
          <Icon name="status-waiting-small" />
          {getPercentageString(pendingClusters)}
          Pending
        </div>
        <div>
          <Icon name="status-failed-small" />
          {getPercentageString(degradedClusters)}
          Degraded
        </div>
      </div>
    </div>
  );
};

export default ClusterStatusGraph;

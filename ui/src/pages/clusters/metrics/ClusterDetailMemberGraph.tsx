import { DoughnutChart, Icon } from "@canonical/react-components";
import type { FC, ReactNode } from "react";
import type { Cluster } from "types/cluster";
import { pluralize, statusCount } from "util/helpers";

interface Props {
  cluster: Cluster;
}

const ClusterDetailMemberGraph: FC<Props> = ({ cluster }: Props) => {
  const online = statusCount(cluster.member_statuses, "Online");
  const offline = statusCount(cluster.member_statuses, "Offline");
  const evacuated = statusCount(cluster.member_statuses, "Evacuated");
  const blocked = statusCount(cluster.member_statuses, "Blocked");

  const total = online + offline + evacuated + blocked;

  const getPercentage = (portion: number): ReactNode => (
    <>
      <b>{portion}</b>{" "}
      {total > 0 ? `(${Math.floor((portion / total) * 100)}%) ` : ""}
    </>
  );

  const segments = [{ color: "#D3E4ED", tooltip: "", value: 0 }];
  if (online > 0) {
    segments.push({
      color: "#0E8420",
      tooltip: "Online",
      value: online,
    });
  }
  if (offline > 0) {
    segments.push({
      color: "#CC7900",
      tooltip: "Offline",
      value: offline,
    });
  }
  if (blocked > 0) {
    segments.push({
      color: "#C7162B",
      tooltip: "Blocked",
      value: blocked,
    });
  }
  if (evacuated > 0) {
    segments.push({
      color: "#24598f",
      tooltip: "Evacuated",
      value: evacuated,
    });
  }

  return (
    <div className="cluster-detail-doughnut-graph">
      <DoughnutChart
        chartID="clusterNode"
        segmentHoverWidth={45}
        segmentThickness={40}
        segments={segments}
        size={150}
      />
      <ul className="doughnut-chart__legend u-no-margin--left">
        <li className="u-no-margin p-heading--5 u-no-padding">
          {total} {pluralize("member", total)}
        </li>
        <li>
          <Icon name="status-succeeded-small" />
          {getPercentage(online)} Online
        </li>
        <li>
          <Icon name="status-waiting-small" />
          {getPercentage(offline)}
          Offline
        </li>
        <li>
          <Icon name="status-in-progress-small" />
          {getPercentage(evacuated)}
          Evacuated
        </li>
        <li>
          <Icon name="status-failed-small" />
          {getPercentage(blocked)}
          Blocked
        </li>
      </ul>
    </div>
  );
};

export default ClusterDetailMemberGraph;

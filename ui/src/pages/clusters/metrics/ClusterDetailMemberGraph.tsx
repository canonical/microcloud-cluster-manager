import { DoughnutChart } from "@canonical/react-components";
import StatusIcon from "components/StatusIcon";
import type { FC, ReactNode } from "react";
import type { Cluster } from "types/cluster";
import { pluralize, statusCount } from "util/helpers";
import { statusColors } from "util/statusColors";

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
      color: statusColors.positive,
      tooltip: "Online",
      value: online,
    });
  }
  if (offline > 0) {
    segments.push({
      color: statusColors.caution,
      tooltip: "Offline",
      value: offline,
    });
  }
  if (blocked > 0) {
    segments.push({
      color: statusColors.negative,
      tooltip: "Blocked",
      value: blocked,
    });
  }
  if (evacuated > 0) {
    segments.push({
      color: statusColors.neutral,
      tooltip: "Evacuated",
      value: evacuated,
    });
  }

  return (
    <div className="cluster-detail-doughnut-graph">
      <DoughnutChart
        chartID="clusterMember"
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
          <StatusIcon tone="positive" />
          {getPercentage(online)} Online
        </li>
        <li>
          <StatusIcon tone="caution" />
          {getPercentage(offline)}
          Offline
        </li>
        <li>
          <StatusIcon tone="neutral" />
          {getPercentage(evacuated)}
          Evacuated
        </li>
        <li>
          <StatusIcon tone="negative" />
          {getPercentage(blocked)}
          Blocked
        </li>
      </ul>
    </div>
  );
};

export default ClusterDetailMemberGraph;

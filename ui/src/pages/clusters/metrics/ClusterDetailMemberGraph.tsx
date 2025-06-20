import { Icon } from "@canonical/react-components";
import DoughnutChart from "components/DoughnutChart";
import { FC, ReactNode } from "react";
import { Cluster } from "types/cluster";
import { pluralize, statusCount } from "util/helpers";

interface Props {
  cluster: Cluster;
}

const ClusterDetailMemberGraph: FC<Props> = ({ cluster }: Props) => {
  const onlineMembers = statusCount(cluster.member_statuses, "Online");
  const offlineMembers = statusCount(cluster.member_statuses, "Offline");
  const evacuatedMembers = statusCount(cluster.member_statuses, "Evacuated");
  const blockedMembers = statusCount(cluster.member_statuses, "Blocked");

  const totalMembers =
    onlineMembers + offlineMembers + evacuatedMembers + blockedMembers;

  function getPercentageString(portion: number): ReactNode {
    return (
      <>
        <b>{portion}</b>{" "}
        {totalMembers > 0
          ? `(${Math.floor((portion / totalMembers) * 100)}%) `
          : ""}
      </>
    );
  }

  return (
    <div className="cluster-detail-doughnut-graph">
      <DoughnutChart
        chartID="clusterNode"
        segmentHoverWidth={45}
        segmentWidth={40}
        segments={[
          {
            color: "#0E8420",
            tooltip: "Online",
            value: onlineMembers,
          },
          {
            color: "#CC7900",
            tooltip: "Offline",
            value: offlineMembers,
          },
          {
            color: "#24598f",
            tooltip: "Evacuated",
            value: evacuatedMembers,
          },
          { color: "#C7162B", tooltip: "Blocked", value: blockedMembers },
        ]}
        size={150}
      />
      <ul className="doughnut-chart__legend u-no-margin--left">
        <li className="u-no-margin p-heading--5 u-no-padding">
          {totalMembers} {pluralize("member", totalMembers)}
        </li>
        <li>
          <Icon name="status-succeeded-small" />
          {getPercentageString(onlineMembers)} Online
        </li>
        <li>
          <Icon name="status-waiting-small" />
          {getPercentageString(offlineMembers)}
          Offline
        </li>
        <li>
          <Icon name="status-in-progress-small" />
          {getPercentageString(evacuatedMembers)}
          Evacuated
        </li>
        <li>
          <Icon name="status-failed-small" />
          {getPercentageString(blockedMembers)}
          Blocked
        </li>
      </ul>
    </div>
  );
};

export default ClusterDetailMemberGraph;

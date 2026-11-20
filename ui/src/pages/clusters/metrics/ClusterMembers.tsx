import { Tooltip } from "@canonical/react-components";
import StatusIcon from "components/StatusIcon";
import type { FC } from "react";
import type { Cluster } from "types/cluster";

interface Props {
  cluster: Cluster;
}

export const ClusterMembers: FC<Props> = ({ cluster }: Props) => {
  const onlineMembers = cluster.member_statuses.find(
    (item) => item.status === "Online",
  ) ?? { count: 0 };
  const offlineMembers = cluster.member_statuses.find(
    (item) => item.status === "Offline",
  ) ?? { count: 0 };
  const evacuatedMembers = cluster.member_statuses.find(
    (item) => item.status === "Evacuated",
  ) ?? { count: 0 };
  const blockedMembers = cluster.member_statuses.find(
    (item) => item.status === "Blocked",
  ) ?? { count: 0 };

  const total =
    onlineMembers.count +
    offlineMembers.count +
    evacuatedMembers.count +
    blockedMembers.count;

  return (
    <Tooltip
      message={
        <div>
          <div>
            <StatusIcon tone="positive" />
            {`Online (${onlineMembers.count})`}
          </div>
          <div>
            <StatusIcon tone="caution" />
            {`Blocked (${blockedMembers.count})`}
          </div>
          <div>
            <StatusIcon tone="negative" />
            {`Offline (${offlineMembers.count})`}
          </div>
          <div>
            <StatusIcon tone="neutral" />
            {`Evacuated (${evacuatedMembers.count})`}
          </div>
        </div>
      }
      position="btm-center"
    >
      <div className="tooltip-toggle">{total}</div>
    </Tooltip>
  );
};

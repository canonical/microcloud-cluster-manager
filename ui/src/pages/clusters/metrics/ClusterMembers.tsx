import { Icon, Tooltip } from "@canonical/react-components";
import { FC } from "react";
import { Cluster } from "types/cluster";

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
            <Icon name="status-succeeded-small" />
            {`Online (${onlineMembers.count})`}
          </div>
          <div>
            <Icon name="status-waiting-small" />
            {`Blocked (${blockedMembers.count})`}
          </div>
          <div>
            <Icon name="status-failed-small" />
            {`Offline (${offlineMembers.count})`}
          </div>
          <div>
            <Icon name="status-queued-small" />
            {`Evacuated (${evacuatedMembers.count})`}
          </div>
        </div>
      }
      positionElementClassName="tooltip"
      position="left"
    >
      <div className="u-pointer">{total}</div>
    </Tooltip>
  );
};

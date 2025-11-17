import type { FC } from "react";
import type { Cluster } from "types/cluster";
import { Icon, Tooltip } from "@canonical/react-components";

interface Props {
  cluster: Cluster;
}

export const ClusterInstances: FC<Props> = ({ cluster }: Props) => {
  const runningInstances = cluster.instance_statuses.find(
    (item) => item.status === "Running",
  ) ?? { count: 0 };
  const stoppedInstances = cluster.instance_statuses.find(
    (item) => item.status === "Stopped",
  ) ?? { count: 0 };
  const frozenInstances = cluster.instance_statuses.find(
    (item) => item.status === "Frozen",
  ) ?? { count: 0 };
  const errorInstances = cluster.instance_statuses.find(
    (item) => item.status === "Error",
  ) ?? { count: 0 };

  return (
    <Tooltip
      message={
        <div>
          <div>
            <Icon name="status-succeeded-small" />
            {`Running (${runningInstances.count})`}
          </div>
          <div>
            <Icon name="status-in-progress-small" />
            {`Frozen (${frozenInstances.count})`}
          </div>
          <div>
            <Icon name="status-failed-small" />
            {`Error (${errorInstances.count})`}
          </div>
          <div>
            <Icon name="status-queued-small" />
            {`Stopped (${stoppedInstances.count})`}
          </div>
        </div>
      }
      position="btm-center"
    >
      <div className="tooltip-toggle">{runningInstances.count}</div>
    </Tooltip>
  );
};

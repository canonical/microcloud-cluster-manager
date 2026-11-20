import type { FC } from "react";
import type { Cluster } from "types/cluster";
import { Tooltip } from "@canonical/react-components";
import StatusIcon from "components/StatusIcon";

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
            <StatusIcon tone="positive" />
            {`Running (${runningInstances.count})`}
          </div>
          <div>
            <StatusIcon tone="information" />
            {`Frozen (${frozenInstances.count})`}
          </div>
          <div>
            <StatusIcon tone="negative" />
            {`Error (${errorInstances.count})`}
          </div>
          <div>
            <StatusIcon tone="neutral" />
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

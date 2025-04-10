import { FC } from "react";
import { Cluster } from "types/cluster";
import MultiMeter from "components/MultiMeter";
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

  const total =
    runningInstances.count +
    stoppedInstances.count +
    frozenInstances.count +
    errorInstances.count;

  const values = [
    {
      amount: runningInstances.count,
      status: "Running",
      color: "#0E8420",
    },
    {
      amount: frozenInstances.count,
      status: "Frozen",
      color: "#24598f",
    },
    {
      amount: errorInstances.count,
      status: "Error",
      color: "#C7162B",
    },
    {
      amount: stoppedInstances.count,
      status: "Stopped",
      color: "#000",
    },
  ];

  const notOkCount = total - runningInstances.count;
  const extra =
    notOkCount > 0
      ? `(${total - runningInstances.count} not running)`
      : "running";

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
      positionElementClassName="tooltip"
      position="btm-center"
    >
      <MultiMeter values={values} text={`${total} ${extra}`} />
    </Tooltip>
  );
};

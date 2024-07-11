import { FC } from "react";
import { Cluster } from "types/cluster";
import Meter from "../Meter";

interface Props {
  cluster: Cluster;
}

export const ClusterInstances: FC<Props> = ({ cluster }: Props) => {
  const runningInstances = cluster.instance_statuses.find(
    (status) => status.status === "Running",
  );
  const runningCount = runningInstances ? runningInstances.count : 0;

  return (
    <Meter
      percentage={(100 / cluster.instance_count) * runningCount || 0}
      text={`${cluster.instance_count} ( ${cluster.instance_count - runningCount} errored)`}
      type="instances"
    />
  );
};

import { FC } from "react";
import { Cluster } from "types/cluster";
import Meter from "../Meter";

interface Props {
  cluster: Cluster;
}

export const ClusterNodes: FC<Props> = ({ cluster }: Props) => {
  const runningMembers = cluster.member_statuses.find(
    (status) => status.status === "Online",
  );
  const activeCount = runningMembers ? runningMembers.count : 0;

  return (
    <Meter
      percentage={(100 / cluster.member_count) * activeCount || 0}
      text={`${cluster.member_count} ( ${cluster.member_count - activeCount} degraded)`}
      type="instances" //Same Design
    />
  );
};

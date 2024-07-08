import { FC } from "react";
import { Cluster } from "types/cluster";
import Meter from "../Meter";
import { humanFileSize } from "util/helpers";

interface Props {
  cluster: Cluster;
}

export const ClusterMemory: FC<Props> = ({ cluster }: Props) => {
  return (
    <Meter
      percentage={
        (100 / cluster.memory_total_amount) * cluster.memory_usage || 0
      }
      text={`${humanFileSize(cluster.memory_usage)} of ${humanFileSize(cluster.memory_total_amount)} memory`}
      type="memory"
    />
  );
};

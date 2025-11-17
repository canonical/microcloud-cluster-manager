import type { FC } from "react";
import type { Cluster } from "types/cluster";
import Meter from "components/Meter";
import { humanFileSize } from "util/helpers";

interface Props {
  cluster: Cluster;
  containerClassname?: string;
}

export const ClusterMemory: FC<Props> = ({
  cluster,
  containerClassname,
}: Props) => {
  return (
    <Meter
      containerClassname={containerClassname || ""}
      percentage={
        (100 / cluster.memory_total_amount) * cluster.memory_usage || 0
      }
      text={`${humanFileSize(cluster.memory_usage)} of ${humanFileSize(cluster.memory_total_amount)}`}
      type="memory"
    />
  );
};

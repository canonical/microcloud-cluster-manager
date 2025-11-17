import type { FC } from "react";
import type { Cluster } from "types/cluster";
import Meter from "components/Meter";
import { humanFileSize } from "util/helpers";

interface Props {
  cluster: Cluster;
  containerClassname?: string;
}

export const ClusterDisk: FC<Props> = ({
  cluster,
  containerClassname,
}: Props) => {
  return (
    <Meter
      containerClassname={containerClassname || ""}
      percentage={(100 / cluster.disk_total_size) * cluster.disk_usage || 0}
      text={`${humanFileSize(cluster.disk_usage)} of ${humanFileSize(cluster.disk_total_size)}`}
      type="disk"
    />
  );
};

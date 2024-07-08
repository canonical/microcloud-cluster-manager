import { FC } from "react";
import { Cluster } from "types/cluster";
import Meter from "../Meter";
import { humanFileSize } from "util/helpers";

interface Props {
  cluster: Cluster;
}

export const ClusterDisk: FC<Props> = ({ cluster }: Props) => {
  return (
    <Meter
      percentage={(100 / cluster.disk_total_size) * cluster.disk_usage || 0}
      text={`${humanFileSize(cluster.disk_usage)} of ${humanFileSize(cluster.disk_total_size)}`}
      type="disk"
    />
  );
};

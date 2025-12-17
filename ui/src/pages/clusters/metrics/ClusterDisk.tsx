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
  const totalUsage = cluster.storage_pool_usages.reduce(
    (acc, pool) => acc + pool.usage,
    0,
  );
  const totalSize = cluster.storage_pool_usages.reduce(
    (acc, pool) => acc + pool.total,
    0,
  );

  return (
    <Meter
      containerClassname={containerClassname || ""}
      percentage={(100 / totalSize) * totalUsage || 0}
      text={`${humanFileSize(totalUsage)} of ${humanFileSize(totalSize)}`}
      type="disk"
    />
  );
};

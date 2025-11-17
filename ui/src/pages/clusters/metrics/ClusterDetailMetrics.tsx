import type { Cluster } from "types/cluster";
import { ClusterMemory } from "./ClusterMemory";
import { ClusterDisk } from "./ClusterDisk";
import type { FC } from "react";

interface Props {
  cluster: Cluster;
}

const ClusterDetailMetrics: FC<Props> = ({ cluster }: Props) => {
  return (
    <div className="cluster-detail-metrics">
      <div className="meter-row">
        <span className="meter-row__title u-no-margin p-heading--5 u-no-padding">
          Total memory
        </span>
        <ClusterMemory
          cluster={cluster}
          containerClassname="meter-row__metrics"
        />
      </div>
      <div className="meter-row">
        <span className="meter-row__title u-no-margin p-heading--5 u-no-padding">
          Total storage
        </span>
        <ClusterDisk
          cluster={cluster}
          containerClassname="meter-row__metrics"
        />
      </div>
    </div>
  );
};

export default ClusterDetailMetrics;

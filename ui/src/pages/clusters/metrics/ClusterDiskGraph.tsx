import PercentileChart from "components/PercentileChart";
import { FC } from "react";
import { Cluster } from "types/cluster";

interface Props {
  clusters: Cluster[];
}

const ClusterDiskGraph: FC<Props> = ({ clusters }) => {
  const diskUsagePercentages = clusters.map(
    (cluster) => cluster.disk_usage / cluster.disk_total_size || 0,
  );

  diskUsagePercentages.sort((a, b) => b - a);

  return (
    <PercentileChart
      title="Disk usage in %"
      barClassName="cluster-disk-bar"
      data={diskUsagePercentages}
      width={200}
      height={90}
    />
  );
};

export default ClusterDiskGraph;

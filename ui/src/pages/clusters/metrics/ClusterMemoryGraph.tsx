import PercentileChart from "components/PercentileChart";
import { FC } from "react";
import { Cluster } from "types/cluster";

interface Props {
  clusters: Cluster[];
}

const ClusterMemoryGraph: FC<Props> = ({ clusters }) => {
  const memoryUsagePercentages = clusters.map(
    (cluster) => cluster.memory_usage / cluster.memory_total_amount || 0,
  );

  memoryUsagePercentages.sort((a, b) => b - a);

  return (
    <PercentileChart
      title="Memory usage"
      barClassName="cluster-memory-bar"
      data={memoryUsagePercentages}
      width={200}
      height={90}
    />
  );
};

export default ClusterMemoryGraph;

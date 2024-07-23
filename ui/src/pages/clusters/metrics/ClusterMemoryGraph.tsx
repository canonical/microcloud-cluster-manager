import { useQuery } from "@tanstack/react-query";
import { fetchClusters } from "api/clusters";
import PercentileChart from "components/PercentileChart";
import { FC } from "react";
import { queryKeys } from "util/queryKeys";

const ClusterMemoryGraph: FC = () => {
  const { data: clusters = [] } = useQuery({
    queryKey: [queryKeys.clusters],
    queryFn: fetchClusters,
  });

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

import { useQuery } from "@tanstack/react-query";
import { fetchClusters } from "api/clusters";
import PercentileChart from "components/PercentileChart";
import { FC } from "react";
import { queryKeys } from "util/queryKeys";

const ClusterDiskGraph: FC = () => {
  const { data: clusters = [] } = useQuery({
    queryKey: [queryKeys.clusters],
    queryFn: fetchClusters,
  });

  const diskUsagePercentages = clusters.map(
    (cluster) => cluster.disk_usage / cluster.disk_total_size || 0,
  );

  diskUsagePercentages.sort((a, b) => b - a);

  return (
    <PercentileChart
      title="Disk usage"
      barClassName="cluster-disk-bar"
      data={diskUsagePercentages}
      width={200}
      height={90}
    />
  );
};

export default ClusterDiskGraph;

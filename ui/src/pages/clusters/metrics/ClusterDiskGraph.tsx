import { useQuery } from "@tanstack/react-query";
import { fetchClusters } from "api/clusters";
import HorizontalDistGraph from "components/HorizontalDistGraph";
import { FC } from "react";
import { queryKeys } from "util/queryKeys";

const GRAPH_COLOR = "#0066CC";

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
    <HorizontalDistGraph
      title="Disk usage"
      color={GRAPH_COLOR}
      data={diskUsagePercentages}
      width={200}
      height={90}
    />
  );
};

export default ClusterDiskGraph;

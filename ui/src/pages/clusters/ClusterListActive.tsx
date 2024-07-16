import { FC } from "react";
import { fetchClusters } from "api/clusters";
import { useQuery } from "@tanstack/react-query";
import { MainTable, TablePagination } from "@canonical/react-components";
import Loader from "components/Loader";
import { Link } from "react-router-dom";
import { ClusterInstances } from "components/meters/ClusterInstances";
import { ClusterCpu } from "components/meters/ClusterCpu";
import { ClusterMemory } from "components/meters/ClusterMemory";
import { ClusterDisk } from "components/meters/ClusterDisk";
import { ClusterNodes } from "components/meters/ClusterNodes";
import ClusterHeartbeat from "components/ClusterHeartbeat";
import { queryKeys } from "util/queryKeys";

const ClusterListActive: FC = () => {
  const { data: clusters = [], isLoading } = useQuery({
    queryKey: [queryKeys.clusters],
    queryFn: fetchClusters,
  });

  const filteredClusters = clusters.filter(
    (cluster) => cluster.status == "ACTIVE",
  );

  const tableHeaders = [
    {
      content: "Cluster Name",
    },
    {
      content: "Last Heartbeat",
    },
    {
      content: "Nodes",
    },
    {
      content: "Instances",
    },
    {
      content: "CPU",
    },
    {
      content: "Memory",
    },
    {
      content: "Disk",
    },
  ];

  const tableRows = filteredClusters.map((cluster) => {
    return {
      columns: [
        {
          content: (
            <Link to={`/ui/cluster/${cluster.name}`}>{cluster.name}</Link>
          ),
        },
        { content: <ClusterHeartbeat cluster={cluster} /> },
        {
          content: <ClusterNodes cluster={cluster} />,
        },
        {
          content: <ClusterInstances cluster={cluster} />,
        },
        {
          content: <ClusterCpu cluster={cluster} />,
        },
        {
          content: <ClusterMemory cluster={cluster} />,
        },
        {
          content: <ClusterDisk cluster={cluster} />,
        },
      ],
    };
  });

  return (
    <div
      role="tabpanel"
      aria-labelledby="Active"
      className="clusterlist-table-container"
    >
      <TablePagination
        data={tableRows}
        id="pagination"
        itemName="cluster"
        className="u-no-margin--top"
        aria-label="Table pagination control"
      >
        <MainTable
          className={"clusterlist-table"}
          responsive={true}
          emptyStateMsg={
            isLoading ? (
              <Loader text="Loading Clusters..." />
            ) : (
              <>No clusters found matching this search.</>
            )
          }
          headers={tableHeaders}
          rows={tableRows}
        />
      </TablePagination>
    </div>
  );
};

export default ClusterListActive;

import { FC } from "react";
import { queryKeys } from "util/queryKeys";
import { fetchClusters } from "api/clusters";
import { useQuery } from "@tanstack/react-query";
import { MainTable, Row, TablePagination } from "@canonical/react-components";
import Loader from "components/Loader";
import BaseLayout from "components/BaseLayout";
import { Link } from "react-router-dom";
import { ClusterInstances } from "components/meters/ClusterInstances";
import { ClusterCpu } from "components/meters/ClusterCpu";
import { ClusterMemory } from "components/meters/ClusterMemory";
import { ClusterDisk } from "components/meters/ClusterDisk";
import { ClusterNodes } from "components/meters/ClusterNodes";
import ClusterHeartbeat from "components/ClusterHeartbeat";

const ClusterList: FC = () => {
  const { data: clusters = [], isLoading } = useQuery({
    queryKey: [queryKeys.clusters],
    queryFn: fetchClusters,
  });

  const tableRows = clusters.map((cluster) => {
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

  return (
    <BaseLayout title={"Clusters"}>
      <Row>
        <div>
          <div className="clusterlist-table-container">
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
                    <Loader text="Loading clusters..." />
                  ) : (
                    <>No instance found matching this search</>
                  )
                }
                headers={tableHeaders}
                rows={tableRows}
              />
            </TablePagination>
          </div>
        </div>
      </Row>
    </BaseLayout>
  );
};

export default ClusterList;

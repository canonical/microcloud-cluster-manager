import { FC } from "react";
import { MainTable, TablePagination } from "@canonical/react-components";
import Loader from "components/Loader";
import { Link } from "react-router-dom";
import { ClusterInstances } from "./metrics/ClusterInstances";
import { ClusterCpu } from "./metrics/ClusterCpu";
import { ClusterMemory } from "./metrics/ClusterMemory";
import { ClusterDisk } from "./metrics/ClusterDisk";
import { ClusterNodes } from "./metrics/ClusterNodes";
import ClusterHeartbeat from "./metrics/ClusterHeartbeat";
import ClusterStatus from "./metrics/ClusterStatus";
import { Cluster } from "types/cluster";

type Props = {
  clusters: Cluster[];
  isLoading: boolean;
};

const ClusterListActive: FC<Props> = ({ clusters, isLoading }) => {
  const filteredClusters = clusters.filter(
    (cluster) => cluster.status == "ACTIVE",
  );

  const tableHeaders = [
    {
      content: "Cluster Name",
    },
    {
      content: "Status",
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
        { content: <ClusterStatus cluster={cluster} /> },
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
        itemName=" active cluster"
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

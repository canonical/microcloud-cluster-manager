import { FC } from "react";
import { MainTable, TablePagination } from "@canonical/react-components";
import Loader from "components/Loader";
import { Link } from "react-router-dom";
import { ClusterInstances } from "./metrics/ClusterInstances";
import { ClusterCpu } from "./metrics/ClusterCpu";
import { ClusterMemory } from "./metrics/ClusterMemory";
import { ClusterDisk } from "./metrics/ClusterDisk";
import { ClusterMembers } from "./metrics/ClusterMembers";
import ClusterHeartbeat from "./metrics/ClusterHeartbeat";
import ClusterStatus from "./metrics/ClusterStatus";
import { Cluster } from "types/cluster";

type Props = {
  clusters: Cluster[];
  isLoading: boolean;
};

const ClusterListActive: FC<Props> = ({ clusters, isLoading }) => {
  const tableHeaders = [
    {
      content: "Cluster Name",
      sortKey: "name",
    },
    {
      content: "Last Heartbeat",
      sortKey: "lastHeartbeat",
    },
    {
      content: "Members",
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
    {
      content: "Status",
      sortKey: "status",
    },
  ];

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
          content: <ClusterMembers cluster={cluster} />,
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
        { content: <ClusterStatus cluster={cluster} /> },
      ],
      sortData: {
        name: cluster.name,
        status: cluster.last_status_update_at,
        lastHeartbeat: cluster.last_status_update_at,
      },
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
          sortable
        />
      </TablePagination>
    </div>
  );
};

export default ClusterListActive;

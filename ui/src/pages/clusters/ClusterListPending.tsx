import { FC } from "react";
import { fetchClusters } from "api/clusters";
import { useQuery } from "@tanstack/react-query";
import { MainTable, TablePagination } from "@canonical/react-components";
import Loader from "components/Loader";
import { queryKeys } from "util/queryKeys";
import PendingClusterActions from "./PendingClusterActions";
import { isoTimeToString } from "util/helpers";

const ClusterListPending: FC = () => {
  const { data: clusters = [], isLoading } = useQuery({
    queryKey: [queryKeys.clusters],
    queryFn: fetchClusters,
  });

  const filteredClusters = clusters.filter(
    (cluster) => cluster.status == "PENDING_APPROVAL",
  );

  const tableHeaders = [
    {
      content: "Cluster Name",
    },
    {
      content: "Join date",
    },
    {
      content: "Actions",
    },
  ];

  const tableRows = filteredClusters.map((cluster) => {
    return {
      columns: [
        {
          content: `${cluster.name}`,
        },
        { content: `${isoTimeToString(cluster.joined_at)}` },
        {
          content: <PendingClusterActions cluster={cluster} />,
        },
      ],
    };
  });

  return (
    <div
      role="tabpanel"
      aria-labelledby="Pending"
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
              <>No pending clusters.</>
            )
          }
          headers={tableHeaders}
          rows={tableRows}
        />
      </TablePagination>
    </div>
  );
};

export default ClusterListPending;

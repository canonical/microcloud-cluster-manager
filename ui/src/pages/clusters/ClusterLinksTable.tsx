import { MainTable, TablePagination } from "@canonical/react-components";
import type { FC } from "react";
import useSortTableData from "context/useSortTableData";
import CreateClusterLinkBtn from "components/cluster-links/CreateClusterLinkBtn";
import getClusterLinkRow, {
  clusterLinkHeaders,
} from "components/cluster-links/ClusterLinkRow";
import type { Cluster, ClusterLink } from "types/cluster";
import { getSourceClusterConnections } from "util/clusterLinks";

interface Props {
  cluster: Cluster;
  clusters: Cluster[];
  links: ClusterLink[];
}

const ClusterLinksTable: FC<Props> = ({ cluster, clusters, links }) => {
  const connections = getSourceClusterConnections(cluster, links, clusters);

  const { rows: sortedRows, updateSort } = useSortTableData({
    rows: connections.map(getClusterLinkRow),
  });

  return (
    <>
      <div className="cluster-links-create">
        <CreateClusterLinkBtn cluster={cluster.name} />
      </div>
      <h2 className="p-heading--4 u-no-padding--top">Cluster link details</h2>
      <TablePagination
        data={sortedRows}
        id="cluster-links-pagination"
        itemName="link"
        className="u-no-margin--top u-no-margin--bottom"
        aria-label="Table pagination control"
        pageLimits={[10, 25, 50]}
        position="below"
      >
        <MainTable
          id="cluster-links-table"
          className="cluster-links-table u-no-margin--bottom"
          responsive
          expanding={false}
          sortable
          headers={clusterLinkHeaders}
          rows={sortedRows}
          onUpdateSort={updateSort}
        />
      </TablePagination>
    </>
  );
};

export default ClusterLinksTable;

import { ScrollableTable, TablePagination } from "@canonical/react-components";
import type { FC } from "react";
import { useState } from "react";
import { useSearchParams } from "react-router-dom";
import useSortTableData from "context/useSortTableData";
import SelectedTableNotification from "components/SelectedTableNotification";
import SelectableMainTable from "components/SelectableMainTable";
import CreateClusterLinkButton from "components/cluster-links/CreateClusterLinkButton";
import DeleteClusterLinksButton from "components/cluster-links/DeleteClusterLinksButton";
import ClusterLinkSearchFilter, {
  getClusterLinkFilters,
} from "components/cluster-links/ClusterLinkSearchFilter";
import getClusterLinkRow, {
  clusterLinkHeaders,
} from "components/cluster-links/ClusterLinkRow";
import type { Cluster, ResolvedClusterLink } from "types/cluster";
import { filterClusterLinks } from "util/clusterLinks";

interface Props {
  cluster: Cluster;
  clusterLinks: ResolvedClusterLink[];
}

const ClusterLinksTable: FC<Props> = ({ cluster, clusterLinks }) => {
  const [selectedNames, setSelectedNames] = useState<string[]>([]);
  const [processingNames, setProcessingNames] = useState<string[]>([]);
  const [searchParams] = useSearchParams();
  const filteredLinks = filterClusterLinks(
    clusterLinks,
    getClusterLinkFilters(searchParams),
  );
  const filteredNames = filteredLinks.map(({ link }) => link.name);
  const validSelectedNames = selectedNames.filter((name) =>
    filteredNames.includes(name),
  );
  const selectedLinks = filteredLinks.filter(({ link }) =>
    validSelectedNames.includes(link.name),
  );

  const { rows: sortedRows, updateSort } = useSortTableData({
    rows: filteredLinks.map(getClusterLinkRow),
  });

  const selectedNotification = validSelectedNames.length > 0 && (
    <SelectedTableNotification
      totalCount={clusterLinks.length}
      itemName="link"
      parentName="cluster"
      selectedNames={validSelectedNames}
      setSelectedNames={setSelectedNames}
      filteredNames={filteredNames}
    />
  );

  return (
    <>
      <div className="cluster-links-create">
        <CreateClusterLinkButton cluster={cluster.name} />
      </div>
      <div className="cluster-links-header">
        <h2 className="p-heading--4 u-no-margin--bottom u-no-padding--top">
          Cluster link details
        </h2>
        <div className="cluster-links-search">
          <ClusterLinkSearchFilter />
        </div>
        {selectedLinks.length > 0 && (
          <div className="cluster-links-bulk-actions">
            <DeleteClusterLinksButton
              cluster={cluster.name}
              links={selectedLinks}
              onStart={() => {
                setProcessingNames(validSelectedNames);
              }}
              onFinish={() => {
                setProcessingNames([]);
                setSelectedNames([]);
              }}
            />
          </div>
        )}
      </div>
      <ScrollableTable
        dependencies={[
          clusterLinks,
          filteredNames.length,
          validSelectedNames.length,
        ]}
        tableId="cluster-links-table"
        belowIds={["status-bar", "cluster-links-pagination"]}
      >
        <TablePagination
          data={sortedRows}
          id="cluster-links-pagination"
          itemName="link"
          className="u-no-margin--top u-no-margin--bottom"
          aria-label="Table pagination control"
          pageLimits={[10, 25, 50]}
          position="below"
          description={selectedNotification}
        >
          <SelectableMainTable
            id="cluster-links-table"
            className="cluster-links-table"
            responsive
            expanding={false}
            sortable
            headers={clusterLinkHeaders}
            emptyStateMsg="No cluster links found matching this search."
            rows={sortedRows}
            onUpdateSort={updateSort}
            selectedNames={validSelectedNames}
            setSelectedNames={setSelectedNames}
            itemName="link"
            parentName="cluster"
            filteredNames={filteredNames}
            disabledNames={processingNames}
            hideContextualMenu
          />
        </TablePagination>
      </ScrollableTable>
    </>
  );
};

export default ClusterLinksTable;

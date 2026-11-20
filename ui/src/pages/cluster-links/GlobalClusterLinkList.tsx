import { sortGroupedRows } from "util/groupedTableSort";
import type { FC } from "react";
import {
  EmptyState,
  Notification,
  Row,
  ScrollableTable,
  TablePagination,
} from "@canonical/react-components";
import { useSearchParams } from "react-router-dom";
import CustomLayout from "components/CustomLayout";
import NotificationRow from "components/NotificationRow";
import PageHeader from "components/PageHeader";
import usePanelParams, { panels } from "context/usePanelParams";
import { useAllClusterLinks } from "context/useClusters";
import { useTableSort } from "context/useSortTableData";
import {
  CLUSTER_LINK_GROUP_SORT_COLUMN,
  filterClusterLinks,
  getClusterLinkId,
  getClusterLinkSortData,
} from "util/clusterLinks";
import CreateClusterLinkPanel from "components/cluster-links/CreateClusterLinkPanel";
import EditClusterLinkPanel from "components/cluster-links/EditClusterLinkPanel";
import ClusterLinkSearchFilter, {
  getClusterLinkFilters,
} from "components/cluster-links/ClusterLinkSearchFilter";
import GroupedClusterLinksTable from "pages/cluster-links/GroupedClusterLinksTable";
import CreateClusterLinkButton from "components/cluster-links/CreateClusterLinkButton";
import ResourceIcon from "components/ResourceIcon";

const GlobalClusterLinkList: FC = () => {
  const panelParams = usePanelParams();
  const [searchParams] = useSearchParams();

  const { data, error, isLoading } = useAllClusterLinks();
  const links = data?.cluster_links ?? [];
  const failedClusters = data?.errors ?? [];
  const hasLoadErrors = !!error || failedClusters.length > 0;
  const showNotificationRow =
    !hasLoadErrors && panelParams.panel !== panels.createClusterLink;

  const filters = getClusterLinkFilters(searchParams);
  const hasFilters = filters.queries.length > 0 || filters.types.length > 0;

  // Sorted before pagination, so that the sort applies to all pages.
  const { sort, updateSort } = useTableSort();
  const sortedLinks = sortGroupedRows(
    filterClusterLinks(links, filters),
    sort.key,
    sort.direction,
    {
      getGroupKey: (resolvedLink) => resolvedLink.source_cluster,
      groupSortColumn: CLUSTER_LINK_GROUP_SORT_COLUMN,
      getRowId: getClusterLinkId,
      getRowValues: getClusterLinkSortData,
    },
  );

  const isEmptyState =
    links.length === 0 &&
    !isLoading &&
    !hasFilters &&
    !error &&
    failedClusters.length === 0;

  return (
    <>
      <CustomLayout
        mainClassName="cluster-link-page"
        header={
          <PageHeader>
            <PageHeader.Left>
              <PageHeader.Title>Cluster links</PageHeader.Title>
              <PageHeader.Search>
                {!isEmptyState && <ClusterLinkSearchFilter />}
              </PageHeader.Search>
            </PageHeader.Left>
            <PageHeader.BaseActions>
              {!isEmptyState && (
                <div className="u-float-right">
                  <CreateClusterLinkButton />
                </div>
              )}
            </PageHeader.BaseActions>
          </PageHeader>
        }
      >
        <Row>
          {hasLoadErrors && (
            <Notification
              severity="negative"
              title="Loading cluster links failed"
            >
              {error && <p className="u-no-margin--bottom">{error.message}</p>}
              {failedClusters.map(({ cluster, error: clusterError }) => (
                <p key={cluster} className="u-no-margin--bottom">
                  {`Unable to load cluster links for ${cluster}: ${clusterError}`}
                </p>
              ))}
            </Notification>
          )}
          {showNotificationRow && <NotificationRow />}
          {isEmptyState ? (
            <div className="u-no-margin--top">
              <EmptyState
                className="empty-state"
                image={
                  <ResourceIcon
                    type="cluster-link"
                    className="empty-state-icon"
                  />
                }
                title="No cluster links found"
              >
                <p>
                  There are no cluster links. Create a cluster link to connect
                  two clusters.
                </p>
                <CreateClusterLinkButton />
              </EmptyState>
            </div>
          ) : (
            <ScrollableTable
              dependencies={[links, failedClusters.length, isLoading]}
              tableId="cluster-link-table"
              belowIds={["status-bar", "pagination"]}
            >
              <TablePagination
                data={sortedLinks}
                id="pagination"
                itemName="cluster link"
                className="u-no-margin--bottom"
                aria-label="Table pagination control"
                pageLimits={[10, 20, 25, 50]}
                position="below"
              >
                <GroupedClusterLinksTable
                  isLoading={isLoading}
                  onUpdateSort={updateSort}
                />
              </TablePagination>
            </ScrollableTable>
          )}
        </Row>
      </CustomLayout>
      {panelParams.panel === panels.createClusterLink && (
        <CreateClusterLinkPanel />
      )}
      {panelParams.panel === panels.editClusterLink && <EditClusterLinkPanel />}
    </>
  );
};

export default GlobalClusterLinkList;

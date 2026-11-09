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
import { useClustersWithLinks } from "context/useClusters";
import useSortTableData from "context/useSortTableData";
import {
  filterClusterLinks,
  getClusterConnections,
  sortGlobalClusterConnections,
} from "util/clusterLinks";
import type { ClusterLinkFilter } from "util/clusterLinks";
import CreateClusterLinkPanel from "pages/clusters/CreateClusterLinkPanel";
import ClusterLinkSearchFilter, {
  LINK_TYPE,
  QUERY,
} from "pages/cluster-links/ClusterLinkSearchFilter";
import ClusterLinkTable from "pages/cluster-links/ClusterLinkTable";
import CreateClusterLinkButton from "pages/cluster-links/actions/CreateClusterLinkButton";
import ResourceIcon from "components/ResourceIcon";

const GlobalClusterLinkList: FC = () => {
  const panelParams = usePanelParams();
  const [searchParams] = useSearchParams();

  const { data: clusters = [], error, isLoading } = useClustersWithLinks();
  const links = getClusterConnections(clusters);
  const failedClusters = clusters.filter(
    (cluster) => cluster.cluster_links_error,
  );
  const hasLoadErrors = !!error || failedClusters.length > 0;
  const showNotificationRow =
    !hasLoadErrors && panelParams.panel !== panels.createClusterLink;

  const filters: ClusterLinkFilter = {
    queries: searchParams.getAll(QUERY),
    types: searchParams.getAll(LINK_TYPE),
  };
  const hasFilters = filters.queries.length > 0 || filters.types.length > 0;

  // Sorted before pagination, so that the sort applies to all pages.
  const { rows: sortedLinks, updateSort } = useSortTableData({
    rows: filterClusterLinks(links, filters),
    sortRows: sortGlobalClusterConnections,
  });

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
              {failedClusters.map(({ name, cluster_links_error }) => (
                <p key={name} className="u-no-margin--bottom">
                  {`Unable to load cluster links for ${name}: ${cluster_links_error}`}
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
                pageLimits={[50]}
                position="below"
              >
                <ClusterLinkTable
                  groupBySource
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
    </>
  );
};

export default GlobalClusterLinkList;

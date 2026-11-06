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
import { useClusterLinks } from "context/useClusterLinks";
import { useClusters } from "context/useClusters";
import { filterClusterLinks } from "util/clusterLinks";
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

  const {
    data: clusters = [],
    error: clustersError,
    isLoading: isClustersLoading,
  } = useClusters();
  const {
    connections: links,
    failedGroups,
    error: linksError,
    isLoading: isLinksLoading,
  } = useClusterLinks(clusters);

  const filters: ClusterLinkFilter = {
    queries: searchParams.getAll(QUERY).map((value) => value.toLowerCase()),
    types: searchParams.getAll(LINK_TYPE),
  };
  const hasFilters = filters.queries.length > 0 || filters.types.length > 0;

  const filteredLinks = filterClusterLinks(links, filters);

  const isLoading = isClustersLoading || isLinksLoading;
  const isEmptyState =
    links.length === 0 &&
    !isLoading &&
    !hasFilters &&
    !clustersError &&
    !linksError &&
    failedGroups.length === 0;

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
          {panelParams.panel !== panels.createClusterLink && (
            <NotificationRow />
          )}
          {clustersError && (
            <Notification severity="negative" title="Loading clusters failed">
              {clustersError.message}
            </Notification>
          )}
          {linksError && (
            <Notification
              severity="negative"
              title="Loading cluster links failed"
            >
              {linksError.message}
            </Notification>
          )}
          {failedGroups.length > 0 && (
            <Notification
              severity="negative"
              title="Loading cluster links failed"
            >
              {failedGroups.map(({ cluster, error }) => (
                <p key={cluster} className="u-no-margin--bottom">
                  {`Unable to load cluster links for ${cluster}: ${error}`}
                </p>
              ))}
            </Notification>
          )}
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
              dependencies={[links, failedGroups.length, isLoading]}
              tableId="cluster-link-table"
              belowIds={["status-bar", "pagination"]}
            >
              <TablePagination
                data={filteredLinks}
                id="pagination"
                itemName="cluster link"
                className="u-no-margin--bottom"
                aria-label="Table pagination control"
                pageLimits={[50]}
                position="below"
              >
                <ClusterLinkTable groupBySource isLoading={isLoading} />
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

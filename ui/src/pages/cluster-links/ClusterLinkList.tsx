import type { FC } from "react";
import { useEffect, useState } from "react";
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
import SelectedTableNotification from "components/SelectedTableNotification";
import usePanelParams, { panels } from "context/usePanelParams";
import { useClusterLinks } from "context/useClusterLinks";
import { useClusters } from "context/useClusters";
import { filterClusterLinks } from "util/clusterLinks";
import type { ClusterLinkFilter } from "util/clusterLinks";
import useSortClusterLinks from "util/useSortClusterLinks";
import CreateClusterLinkPanel from "pages/clusters/CreateClusterLinkPanel";
import ClusterLinkSearchFilter, {
  LINK_TYPE,
  QUERY,
} from "pages/cluster-links/ClusterLinkSearchFilter";
import ClusterLinkTable from "pages/cluster-links/ClusterLinkTable";
import CreateClusterLinkButton from "pages/cluster-links/actions/CreateClusterLinkButton";
import DeleteClusterLinksButton from "pages/cluster-links/actions/DeleteClusterLinksButton";
import ResourceIcon from "components/ResourceIcon";

const ClusterLinkList: FC = () => {
  const panelParams = usePanelParams();
  const [searchParams] = useSearchParams();
  const [selectedIds, setSelectedIds] = useState<string[]>([]);
  const [processingIds, setProcessingIds] = useState<string[]>([]);

  const {
    data: clusters = [],
    error: clustersError,
    isLoading: isClustersLoading,
  } = useClusters();
  const {
    connections: links,
    failedClusters,
    isLoading: isLinksLoading,
  } = useClusterLinks(clusters);

  const filters: ClusterLinkFilter = {
    queries: searchParams.getAll(QUERY).map((value) => value.toLowerCase()),
    types: searchParams.getAll(LINK_TYPE),
  };
  const hasFilters = filters.queries.length > 0 || filters.types.length > 0;

  const filteredLinks = filterClusterLinks(links, filters);
  const { rows: sortedLinks, ...sort } = useSortClusterLinks({
    rows: filteredLinks,
    groupBySource: true,
  });

  useEffect(() => {
    const validIds = new Set(links.map((entry) => entry.id));
    const validSelections = selectedIds.filter((id) => validIds.has(id));
    if (validSelections.length !== selectedIds.length) {
      setSelectedIds(validSelections);
    }
  }, [links]);

  const isLoading = isClustersLoading || isLinksLoading;
  const isEmptyState =
    links.length === 0 &&
    !isLoading &&
    !hasFilters &&
    !clustersError &&
    failedClusters.length === 0;
  const hasSelection = selectedIds.length > 0;
  const selectedLinks = links.filter((entry) => selectedIds.includes(entry.id));

  return (
    <>
      <CustomLayout
        header={
          <PageHeader>
            <PageHeader.Left>
              <PageHeader.Title>Cluster links</PageHeader.Title>
              <PageHeader.Search>
                {!isEmptyState && !hasSelection && <ClusterLinkSearchFilter />}
              </PageHeader.Search>
            </PageHeader.Left>
            <PageHeader.BaseActions>
              {hasSelection && (
                <div className="u-float-right">
                  <DeleteClusterLinksButton
                    links={selectedLinks}
                    onStart={() => {
                      setProcessingIds(selectedIds);
                    }}
                    onFinish={() => {
                      setProcessingIds([]);
                    }}
                  />
                </div>
              )}
              {!hasSelection && !isEmptyState && (
                <div className="u-float-right">
                  <CreateClusterLinkButton />
                </div>
              )}
            </PageHeader.BaseActions>
          </PageHeader>
        }
      >
        <Row>
          <NotificationRow />
          {clustersError && (
            <Notification severity="negative" title="Loading clusters failed">
              {clustersError.message}
            </Notification>
          )}
          {failedClusters.length > 0 && (
            <Notification
              severity="negative"
              title="Loading cluster links failed"
            >
              Unable to load cluster links for{" "}
              {failedClusters.map((cluster) => cluster.name).join(", ")}.
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
              dependencies={[
                links,
                failedClusters.length,
                isLoading,
                selectedIds.length,
              ]}
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
                description={
                  hasSelection && (
                    <SelectedTableNotification
                      totalCount={links.length}
                      itemName="cluster link"
                      selectedNames={selectedIds}
                      setSelectedNames={setSelectedIds}
                      filteredNames={filteredLinks.map((entry) => entry.id)}
                      hideActions={processingIds.length > 0}
                    />
                  )
                }
              >
                <ClusterLinkTable
                  {...sort}
                  groupBySource
                  isLoading={isLoading}
                  processingIds={processingIds}
                  selectedIds={selectedIds}
                  setSelectedIds={setSelectedIds}
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

export default ClusterLinkList;

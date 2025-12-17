import type { FC } from "react";
import { useEffect, useState } from "react";
import { Row, usePortal } from "@canonical/react-components";
import { useLocation, useParams, useSearchParams } from "react-router-dom";
import TabLinks from "components/TabLinks";
import ClusterListTokens from "./ClusterListTokens";
import ClusterListActive from "./ClusterListActive";
import NotificationRow from "components/NotificationRow";
import ClusterSearchFilter, {
  INSTANCE_STATUS,
  MEMORY_USAGE,
  NODE_STATUS,
  POOL_USAGE,
  QUERY,
} from "pages/clusters/ClusterSearchFilter";
import CustomLayout from "components/CustomLayout";
import PageHeader from "components/PageHeader";
import { useQuery } from "@tanstack/react-query";
import { queryKeys } from "util/queryKeys";
import { fetchClusters } from "api/clusters";
import type { ClusterFilters } from "util/clusterFilter";
import {
  toNumericUsagePercentiles,
  hasAllMatchingStatuses,
  hasAllUsagePercentileBands,
} from "util/clusterFilter";
import type {
  ClusterInstanceStatus,
  ClusterNodeStatus,
  ClusterPercentiles,
} from "types/cluster";
import EnrollClusterModal from "pages/clusters/EnrollClusterModal";
import type { Location } from "react-router-dom";
import { fetchTokens } from "api/tokens";
import usePanelParams, { panels } from "context/usePanelParams";
import EnrollClusterPanel from "pages/clusters/EnrollClusterPanel";
import BulkRemoveClusterButton from "pages/clusters/actions/BulkRemoveClusterButton";
import BulkRevokeTokenButton from "pages/clusters/actions/BulkRevokeTokenButton";
import EnrollClusterButton from "pages/clusters/actions/EnrollClusterButton";
import ConfigureClusterPanel from "pages/clusters/ConfigureClusterPanel";
import BulkConfigureClusterButton from "pages/clusters/actions/BulkConfigureClusterButton";
import BulkConfigureClusterPanel from "pages/clusters/BulkConfigureClusterPanel";

interface ClusterToken {
  name: string;
  token: string;
  expiry: string;
}

export interface TokenState {
  createdCluster?: ClusterToken;
}

const ClusterList: FC = () => {
  const [processingNames, setProcessingNames] = useState<string[]>([]);
  const [selectedNames, setSelectedNames] = useState<string[]>([]);
  const panelParams = usePanelParams();
  const location = useLocation() as Location<TokenState>;
  const { openPortal, closePortal, isOpen, Portal } = usePortal({
    programmaticallyOpen: true,
  });
  const { activeTab } = useParams<{
    activeTab?: string;
  }>();

  const [searchParams] = useSearchParams();

  const tabs: string[] = ["Active", "Tokens"];

  const { data: clusters = [], isLoading } = useQuery({
    queryKey: [queryKeys.clusters],
    queryFn: fetchClusters,
    enabled: !activeTab,
  });

  const { data: tokens = [] } = useQuery({
    queryKey: [queryKeys.tokens],
    queryFn: fetchTokens,
    enabled: activeTab === "tokens",
  });

  const createdCluster = location.state?.createdCluster;
  useEffect(() => {
    if (createdCluster) {
      openPortal();
    }
  }, [createdCluster]);

  useEffect(() => {
    const validNames = new Set(clusters.map((cluster) => cluster.name));
    const validSelections = selectedNames.filter((name) =>
      validNames.has(name),
    );
    if (validSelections.length !== selectedNames.length && !activeTab) {
      setSelectedNames(validSelections);
    }
  }, [clusters, activeTab]);

  const filters: ClusterFilters = {
    queries: searchParams.getAll(QUERY),
    instanceStatuses: searchParams.getAll(
      INSTANCE_STATUS,
    ) as ClusterInstanceStatus[],
    nodeStatuses: searchParams.getAll(NODE_STATUS) as ClusterNodeStatus[],
    memoryUsage: toNumericUsagePercentiles(
      searchParams.getAll(MEMORY_USAGE),
    ) as ClusterPercentiles[],
    storagePoolUsage: toNumericUsagePercentiles(
      searchParams.getAll(POOL_USAGE),
    ) as ClusterPercentiles[],
  };

  const filteredClusters = clusters.filter((item) => {
    if (
      //Query search by name
      !filters.queries.every((q) => item.name.toLowerCase().includes(q))
    ) {
      return false;
    }

    if (
      //Search by Clusters with more than one instance having a particular status
      filters.instanceStatuses.length > 0 &&
      !hasAllMatchingStatuses(filters.instanceStatuses, item.instance_statuses)
    ) {
      return false;
    }

    if (
      //Search by Clusters with more than one node having a particular status
      filters.nodeStatuses.length > 0 &&
      !hasAllMatchingStatuses(filters.nodeStatuses, item.member_statuses)
    ) {
      return false;
    }

    if (
      //Search by Cluster memory usage
      filters.memoryUsage.length > 0 &&
      !hasAllUsagePercentileBands(
        filters.memoryUsage,
        item.memory_usage,
        item.memory_total_amount,
      )
    ) {
      return false;
    }

    //Search by Storage pool usage
    const maxStorageFilter = Math.max(...filters.storagePoolUsage, 0);
    const maxStorageUsage = Math.max(
      ...item.storage_pool_usages.map(
        (usage) => ((100 / usage.total) * usage.usage) / 100 || 0,
      ),
      0,
    );
    if (
      filters.storagePoolUsage.length > 0 &&
      maxStorageFilter > maxStorageUsage
    ) {
      return false;
    }

    return true;
  });

  const isEmptyState =
    (clusters.length === 0 && !activeTab) ||
    (activeTab === "tokens" && tokens.length === 0);
  const hasSearchInput =
    !activeTab && !isEmptyState && selectedNames.length === 0;
  const hasSelectedClusters = !activeTab && selectedNames.length > 0;
  const hasSelectedTokens = activeTab && selectedNames.length > 0;
  const hasCreateInHeader =
    !hasSelectedClusters && !hasSelectedTokens && !isEmptyState;

  return (
    <>
      <CustomLayout
        header={
          <PageHeader>
            <PageHeader.Left>
              <PageHeader.Title>Clusters</PageHeader.Title>

              <PageHeader.Search>
                {hasSearchInput && <ClusterSearchFilter />}
                {hasSelectedClusters && (
                  <>
                    <BulkConfigureClusterButton
                      key="configure"
                      clusterNames={selectedNames}
                      onStart={() => {
                        setProcessingNames(selectedNames);
                      }}
                      onFinish={() => {
                        setProcessingNames([]);
                      }}
                    />
                    <BulkRemoveClusterButton
                      key="remove"
                      clusterNames={selectedNames}
                      onStart={() => {
                        setProcessingNames(selectedNames);
                      }}
                      onFinish={() => {
                        setProcessingNames([]);
                      }}
                    />
                  </>
                )}
                {hasSelectedTokens && (
                  <BulkRevokeTokenButton
                    clusterNames={selectedNames}
                    onStart={() => {
                      setProcessingNames(selectedNames);
                    }}
                    onFinish={() => {
                      setProcessingNames([]);
                    }}
                  />
                )}
              </PageHeader.Search>
            </PageHeader.Left>

            <PageHeader.BaseActions>
              {hasCreateInHeader && (
                <EnrollClusterButton className="u-float-right" />
              )}
            </PageHeader.BaseActions>
          </PageHeader>
        }
      >
        <Row>
          <TabLinks tabs={tabs} activeTab={activeTab} tabUrl="/ui/clusters" />
          <NotificationRow />
          <div>
            {!activeTab && (
              <ClusterListActive
                clusters={filteredClusters}
                isEmptyState={isEmptyState}
                isLoading={isLoading}
                processingNames={processingNames}
                selectedNames={selectedNames}
                setSelectedNames={setSelectedNames}
              />
            )}
            {activeTab === "tokens" && (
              <ClusterListTokens
                selectedNames={selectedNames}
                setSelectedNames={setSelectedNames}
                processingNames={processingNames}
              />
            )}
          </div>
        </Row>
      </CustomLayout>

      {panelParams.panel === panels.enrollCluster && <EnrollClusterPanel />}
      {panelParams.panel === panels.configureCluster && (
        <ConfigureClusterPanel />
      )}
      {panelParams.panel === panels.bulkConfigureCluster && (
        <BulkConfigureClusterPanel />
      )}

      {isOpen && createdCluster && (
        <Portal>
          <EnrollClusterModal
            onClose={closePortal}
            token={createdCluster.token}
            name={createdCluster.name}
            expiry={createdCluster.expiry}
          />
        </Portal>
      )}
    </>
  );
};

export default ClusterList;

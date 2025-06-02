import { FC } from "react";
import { Row } from "@canonical/react-components";
import { useParams, useSearchParams } from "react-router-dom";
import TabLinks from "components/TabLinks";
import ClusterListTokens from "./ClusterListTokens";
import ClusterListActive from "./ClusterListActive";
import AddClusterButton from "./AddClusterButton";
import NotificationRow from "components/NotificationRow";
import ClusterSearchFilter from "pages/clusters/ClusterSearchFilter";
import CustomLayout from "components/CustomLayout";
import PageHeader from "components/PageHeader";
import { useQuery } from "@tanstack/react-query";
import { queryKeys } from "util/queryKeys";
import { fetchClusters } from "api/clusters";
import {
  ClusterFilters,
  toNumericUsagePercentiles,
  hasAllMatchingStatuses,
  hasAllUsagePercentileBands,
} from "util/clusterFilter";
import {
  ClusterInstanceStatus,
  ClusterNodeStatus,
  ClusterPercentiles,
} from "types/cluster";

const ClusterList: FC = () => {
  const { activeTab } = useParams<{
    activeTab?: string;
  }>();

  const [searchParams] = useSearchParams();

  const tabs: string[] = ["Active", "Tokens"];

  const { data: clusters = [], isLoading } = useQuery({
    queryKey: [queryKeys.clusters],
    queryFn: fetchClusters,
  });

  const filters: ClusterFilters = {
    queries: searchParams.getAll("query"),
    instanceStatuses: searchParams.getAll(
      "instance-status",
    ) as ClusterInstanceStatus[],
    nodeStatuses: searchParams.getAll("node-status") as ClusterNodeStatus[],
    memoryUsage: toNumericUsagePercentiles(
      searchParams.getAll("memory-usage"),
    ) as ClusterPercentiles[],
    diskUsage: toNumericUsagePercentiles(
      searchParams.getAll("disk-usage"),
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

    if (
      //Search by Cluster disk usage
      filters.diskUsage.length > 0 &&
      !hasAllUsagePercentileBands(
        filters.diskUsage,
        item.disk_usage,
        item.disk_total_size,
      )
    ) {
      return false;
    }

    return true;
  });

  return (
    <CustomLayout
      header={
        <PageHeader>
          <PageHeader.Left>
            <PageHeader.Title>Clusters</PageHeader.Title>

            <PageHeader.Search>
              {!activeTab && <ClusterSearchFilter />}
            </PageHeader.Search>
          </PageHeader.Left>

          <PageHeader.BaseActions>
            <AddClusterButton />
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
              isLoading={isLoading}
            />
          )}
          {activeTab === "tokens" && <ClusterListTokens />}
        </div>
      </Row>
    </CustomLayout>
  );
};

export default ClusterList;

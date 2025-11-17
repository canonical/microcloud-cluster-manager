import type {
  ClusterInstanceStatus,
  ClusterNodeStatus,
  ClusterPercentiles,
  StatusDistribution,
} from "types/cluster";

export interface ClusterFilters {
  queries: string[];
  instanceStatuses: ClusterInstanceStatus[];
  nodeStatuses: ClusterNodeStatus[];
  memoryUsage: number[];
  diskUsage: number[];
}

export const instanceStatuses: ClusterInstanceStatus[] = [
  "Running",
  "Stopped",
  "Frozen",
  "Error",
];

export const nodeStatuses: ClusterNodeStatus[] = [
  "Online",
  "Blocked",
  "Offline",
  "Evacuated",
];

export const usagePercentiles = [">50%", ">75%", ">80%", ">90%"];

export const toNumericUsagePercentiles = (percentiles: string[]): number[] => {
  const enrichedUsages = [] as ClusterPercentiles[];

  if (percentiles.includes(">50%")) {
    enrichedUsages.push(0.5);
  }
  if (percentiles.includes(">75%")) {
    enrichedUsages.push(0.75);
  }
  if (percentiles.includes(">80%")) {
    enrichedUsages.push(0.8);
  }
  if (percentiles.includes(">90%")) {
    enrichedUsages.push(0.9);
  }

  return enrichedUsages;
};

export const hasAllMatchingStatuses = (
  filteredStatuses: string[],
  itemStatuses: StatusDistribution[],
) => {
  const filterStatusCountLookup: { [status: string]: number } = {};

  itemStatuses.forEach((item) => {
    filterStatusCountLookup[item.status] = item.count;
  });

  return filteredStatuses.every(
    (status) => filterStatusCountLookup[status] > 0,
  );
};

export const hasAllUsagePercentileBands = (
  usagePercentileBands: number[],
  use: number,
  total: number,
) => {
  const usagePercentage = use / total;

  return usagePercentileBands.every((num) => usagePercentage >= num);
};

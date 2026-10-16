import { useState } from "react";
import type {
  ClusterConnection,
  ClusterLinkSortDirection,
  ClusterLinkSortKey,
} from "util/clusterLinks";
import { sortClusterLinks } from "util/clusterLinks";

interface Props {
  rows: ClusterConnection[];
  groupBySource?: boolean;
}

export interface ClusterLinkSort {
  sortKey: ClusterLinkSortKey | null;
  sortDirection: ClusterLinkSortDirection;
  updateSort: (key: ClusterLinkSortKey) => void;
}

interface Result extends ClusterLinkSort {
  rows: ClusterConnection[];
}

const useSortClusterLinks = ({
  rows,
  groupBySource = false,
}: Props): Result => {
  const [sortKey, setSortKey] = useState<ClusterLinkSortKey | null>(null);
  const [sortDirection, setSortDirection] =
    useState<ClusterLinkSortDirection>("ascending");

  const updateSort = (key: ClusterLinkSortKey): void => {
    if (sortKey !== key) {
      setSortKey(key);
      setSortDirection("ascending");
    } else if (sortDirection === "ascending") {
      setSortDirection("descending");
    } else {
      setSortKey(null);
      setSortDirection("ascending");
    }
  };

  return {
    rows: sortClusterLinks(rows, sortKey, sortDirection, groupBySource),
    sortKey,
    sortDirection,
    updateSort,
  };
};

export default useSortClusterLinks;

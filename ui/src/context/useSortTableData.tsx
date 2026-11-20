import { useState } from "react";
import type { MainTableRow } from "@canonical/react-components/dist/components/MainTable/MainTable";

type SortDirection = "ascending" | "descending";

interface SortOptions {
  defaultSort?: string;
  defaultSortDirection?: SortDirection;
}

interface TableSort {
  sort: { key?: string | null; direction: SortDirection };
  updateSort: (sortKey?: string | null) => void;
}

interface TableRowsProps extends SortOptions {
  rows: MainTableRow[];
}

interface SortTableData {
  rows: MainTableRow[];
  updateSort: TableSort["updateSort"];
}

// Same comparison as the MainTable default, so that MainTable keeps the order when it sorts the rows again.
export const sortTableRows = (
  rows: MainTableRow[],
  sortKey: string | null | undefined,
  direction: SortDirection,
): MainTableRow[] => {
  if (!sortKey) {
    return rows;
  }

  return [...rows].sort((a, b) => {
    if (!a.sortData || !b.sortData) {
      return 0;
    }

    const aValue = a.sortData[sortKey] as string | number;
    const bValue = b.sortData[sortKey] as string | number;

    if (aValue > bValue) {
      return direction === "ascending" ? 1 : -1;
    }

    if (aValue < bValue) {
      return direction === "ascending" ? -1 : 1;
    }

    return 0;
  });
};

// Match MainTable's cycle: unsorted -> ascending -> descending -> unsorted.
export const useTableSort = ({
  defaultSort,
  defaultSortDirection = "descending",
}: SortOptions = {}): TableSort => {
  const [sort, setSort] = useState<{
    key?: string | null;
    direction: SortDirection;
  }>({ key: defaultSort, direction: defaultSortDirection });

  const updateSort = (sortKey?: string | null) => {
    setSort((current) => ({
      key: sortKey,
      direction:
        sortKey && sortKey === current.key ? "descending" : "ascending",
    }));
  };

  return { sort, updateSort };
};

// Sort the full set of rows before TablePagination passes the current page
// to MainTable. Connect updateSort to MainTable's onUpdateSort.
const useSortTableData = ({
  rows,
  ...options
}: TableRowsProps): SortTableData => {
  const { sort, updateSort } = useTableSort(options);
  return {
    rows: sortTableRows(rows, sort.key, sort.direction),
    updateSort,
  };
};

export default useSortTableData;

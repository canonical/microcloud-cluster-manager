import { useState } from "react";
import type { MainTableRow } from "@canonical/react-components/dist/components/MainTable/MainTable";

type SortDirection = "ascending" | "descending";

interface Props {
  rows: MainTableRow[];
  defaultSort?: string;
  defaultSortDirection?: SortDirection;
}

interface SortTableData {
  rows: MainTableRow[];
  updateSort: (sortKey?: string | null) => void;
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

// MainTable only sorts the rows it is given. Inside TablePagination those are the rows of the current page, so the
// sort order would only apply within a page. This hook sorts all rows before they are paginated. Pass the returned
// rows to TablePagination, and updateSort to MainTable as onUpdateSort. updateSort follows the same cycle as the
// MainTable headers: unsorted -> ascending -> descending -> unsorted.
const useSortTableData = ({
  rows,
  defaultSort,
  defaultSortDirection = "descending",
}: Props): SortTableData => {
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

  return {
    rows: sortTableRows(rows, sort.key, sort.direction),
    updateSort,
  };
};

export default useSortTableData;

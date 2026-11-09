import { useState } from "react";
import type { MainTableRow } from "@canonical/react-components/dist/components/MainTable/MainTable";

type SortDirection = "ascending" | "descending";

type SortRows<T> = (
  rows: T[],
  sortKey: string | null | undefined,
  direction: SortDirection,
) => T[];

interface Props<T> {
  rows: T[];
  // Sorts the rows. It has to use the same order as the sortFunction of the MainTable.
  sortRows: SortRows<T>;
  defaultSort?: string;
  defaultSortDirection?: SortDirection;
}

interface SortTableData<T> {
  rows: T[];
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
//
// Use sortTableRows to sort MainTable rows by their sort data.
const useSortTableData = <T,>({
  rows,
  sortRows,
  defaultSort,
  defaultSortDirection = "descending",
}: Props<T>): SortTableData<T> => {
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
    rows: sortRows(rows, sort.key, sort.direction),
    updateSort,
  };
};

export default useSortTableData;

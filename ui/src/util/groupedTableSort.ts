type SortDirection = "ascending" | "descending";

type SortValues = Record<string, string | number>;

export interface GroupedRowSortData extends Record<string, unknown> {
  groupIndex: number;
  groupKey: string;
  isGroupHeader: boolean;
  id: string;
  values: SortValues;
}

interface GroupedSortOptions<T> {
  getGroupKey: (row: T) => string;
  groupSortColumn: string;
  getRowId: (row: T) => string;
  getRowValues: (row: T) => SortValues;
}

// Keep group headers above their rows. Sort either the groups or the rows
// within each group, leaving the other order intact.
export const compareGroupedRows = (
  a: GroupedRowSortData,
  b: GroupedRowSortData,
  column: string,
  direction: SortDirection,
  groupSortColumn: string,
): -1 | 0 | 1 => {
  const ascending = direction === "ascending";

  if (a.groupKey !== b.groupKey) {
    if (column === groupSortColumn) {
      const comparison = a.groupKey.localeCompare(b.groupKey);
      if (comparison !== 0) {
        return comparison < 0 === ascending ? -1 : 1;
      }
    }
    return a.groupIndex < b.groupIndex ? -1 : 1;
  }

  if (a.isGroupHeader !== b.isGroupHeader) {
    return a.isGroupHeader ? -1 : 1;
  }

  if (column === groupSortColumn) {
    return 0;
  }

  const aValue = a.values[column];
  const bValue = b.values[column];

  if (aValue < bValue) return ascending ? -1 : 1;
  if (aValue > bValue) return ascending ? 1 : -1;
  if (a.id < b.id) return -1;
  if (a.id > b.id) return 1;
  return 0;
};

// Sort the complete list before the table receives the current page.
export const sortGroupedRows = <T>(
  rows: T[],
  column: string | null | undefined,
  direction: SortDirection,
  {
    getGroupKey,
    groupSortColumn,
    getRowId,
    getRowValues,
  }: GroupedSortOptions<T>,
): T[] => {
  if (!column) {
    return rows;
  }

  const groupIndexes = new Map<string, number>();
  const rowsWithSortData = rows.map((row) => {
    const groupKey = getGroupKey(row);
    const groupIndex = groupIndexes.get(groupKey) ?? groupIndexes.size;
    groupIndexes.set(groupKey, groupIndex);

    const sortData: GroupedRowSortData = {
      groupIndex,
      groupKey,
      isGroupHeader: false,
      id: getRowId(row),
      values: getRowValues(row),
    };
    return { row, sortData };
  });

  return rowsWithSortData
    .sort((a, b) =>
      compareGroupedRows(
        a.sortData,
        b.sortData,
        column,
        direction,
        groupSortColumn,
      ),
    )
    .map(({ row }) => row);
};

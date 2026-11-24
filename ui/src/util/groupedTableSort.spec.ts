import { compareGroupedRows, sortGroupedRows } from "./groupedTableSort";
import type { GroupedRowSortData } from "./groupedTableSort";

describe("compareGroupedRows", () => {
  const groupRow = (group: number, groupKey: string): GroupedRowSortData => ({
    groupIndex: group,
    groupKey,
    id: `group-${groupKey}`,
    isGroupHeader: true,
    values: {},
  });
  const linkRow = (
    group: number,
    groupKey: string,
    id: string,
    values: GroupedRowSortData["values"] = {},
  ): GroupedRowSortData => ({
    ...groupRow(group, groupKey),
    isGroupHeader: false,
    id,
    values,
  });

  it("sorts link rows within their group without moving rows between groups", () => {
    const rows = [
      linkRow(0, "one", "one/b", { name: "b" }),
      linkRow(0, "one", "one/a", { name: "a" }),
      linkRow(1, "two", "two/c", { name: "c" }),
    ];

    const sorted = [...rows].sort((a, b) =>
      compareGroupedRows(a, b, "name", "ascending", "department"),
    );

    expect(sorted.map((row) => row.id)).toStrictEqual([
      "one/a",
      "one/b",
      "two/c",
    ]);
  });

  it("applies the sort direction within the group", () => {
    const rows = [
      linkRow(0, "one", "one/a", { name: "a" }),
      linkRow(0, "one", "one/b", { name: "b" }),
    ];

    const sorted = [...rows].sort((a, b) =>
      compareGroupedRows(a, b, "name", "descending", "department"),
    );

    expect(sorted.map((row) => row.id)).toStrictEqual(["one/b", "one/a"]);
  });

  it("keeps the group header row first within its group", () => {
    const rows = [
      linkRow(0, "one", "one/a", { name: "a" }),
      groupRow(0, "one"),
    ];

    const sorted = [...rows].sort((a, b) =>
      compareGroupedRows(a, b, "name", "descending", "department"),
    );

    expect(sorted.map((row) => row.id)).toStrictEqual(["group-one", "one/a"]);
  });

  it("reorders the groups when sorting by group and keeps link order", () => {
    const rows = [
      groupRow(0, "z-source"),
      linkRow(0, "z-source", "z-source/b", { name: "b" }),
      linkRow(0, "z-source", "z-source/a", { name: "a" }),
      groupRow(1, "a-source"),
      linkRow(1, "a-source", "a-source/c", { name: "c" }),
    ];

    const sorted = [...rows].sort((a, b) =>
      compareGroupedRows(a, b, "department", "ascending", "department"),
    );

    expect(sorted.map((row) => row.id)).toStrictEqual([
      "group-a-source",
      "a-source/c",
      "group-z-source",
      "z-source/b",
      "z-source/a",
    ]);
  });

  it("breaks value ties by row id regardless of direction", () => {
    const rows = [
      linkRow(0, "one", "one/b", { usedBy: 1 }),
      linkRow(0, "one", "one/a", { usedBy: 1 }),
    ];

    for (const direction of ["ascending", "descending"] as const) {
      const sorted = [...rows].sort((a, b) =>
        compareGroupedRows(a, b, "usedBy", direction, "department"),
      );
      expect(sorted.map((row) => row.id)).toStrictEqual(["one/a", "one/b"]);
    }
  });
});

describe("sortGroupedRows", () => {
  const rows = [
    { department: "one", id: "one/b", name: "b" },
    { department: "one", id: "one/a", name: "a" },
    { department: "two", id: "two/d", name: "d" },
    { department: "two", id: "two/c", name: "c" },
  ];
  const sort = (column: string | null, direction: "ascending" | "descending") =>
    sortGroupedRows(rows, column, direction, {
      getGroupKey: (row) => row.department,
      groupSortColumn: "department",
      getRowId: (row) => row.id,
      getRowValues: (row) => ({ name: row.name }),
    });
  const ids = (entries: typeof rows) => entries.map((row) => row.id);

  it("keeps the order without a sort key", () => {
    expect(sort(null, "ascending")).toBe(rows);
  });

  it("sorts rows within their group and keeps the group order", () => {
    expect(ids(sort("name", "ascending"))).toStrictEqual([
      "one/a",
      "one/b",
      "two/c",
      "two/d",
    ]);
    expect(ids(sort("name", "descending"))).toStrictEqual([
      "one/b",
      "one/a",
      "two/d",
      "two/c",
    ]);
  });

  it("reorders groups on the configured column and keeps the row order", () => {
    expect(ids(sort("department", "descending"))).toStrictEqual([
      "two/d",
      "two/c",
      "one/b",
      "one/a",
    ]);
  });

  it("does not change the original rows", () => {
    const before = ids(rows);
    sort("name", "ascending");
    expect(ids(rows)).toStrictEqual(before);
  });
});

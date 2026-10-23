import type { FC } from "react";
import { act, useState } from "react";
import { createRoot } from "react-dom/client";
import { TablePagination } from "@canonical/react-components";
import SelectableMainTable from "components/SelectableMainTable";
import useSortTableData, { sortTableRows } from "./useSortTableData";

// 60 names, more than the default page size of 50, in a shuffled order. The first and last names alphabetically are
// at the end, so they are not on the first page unless all rows are sorted before pagination.
const names = [
  ...Array.from(
    { length: 58 },
    (_, i) => `cluster-${String(((i * 37) % 58) + 1).padStart(2, "0")}`,
  ),
  "cluster-59",
  "cluster-00",
];

const toRow = (name: string) => ({
  key: name,
  name,
  columns: [{ content: name, role: "cell" }],
  sortData: { name },
});

const Harness: FC = () => {
  const [selectedNames, setSelectedNames] = useState<string[]>([]);
  const { rows, updateSort } = useSortTableData({ rows: names.map(toRow) });

  return (
    <>
      <div id="selection">{selectedNames.join(",")}</div>
      <TablePagination data={rows} id="pagination" itemName="cluster">
        <SelectableMainTable
          headers={[{ content: "Name", sortKey: "name" }]}
          rows={rows}
          sortable
          onUpdateSort={updateSort}
          selectedNames={selectedNames}
          setSelectedNames={setSelectedNames}
          itemName="cluster"
          parentName=""
          filteredNames={names}
          disabledNames={[]}
        />
      </TablePagination>
    </>
  );
};

const click = (element: Element | null | undefined, shiftKey = false) => {
  act(() => {
    element?.dispatchEvent(
      new MouseEvent("click", { bubbles: true, shiftKey }),
    );
  });
};

describe("sortTableRows", () => {
  it("sorts by the given key and direction", () => {
    const rows = ["b", "c", "a"].map(toRow);

    expect(sortTableRows(rows, "name", "ascending").map((r) => r.name)).toEqual(
      ["a", "b", "c"],
    );
    expect(
      sortTableRows(rows, "name", "descending").map((r) => r.name),
    ).toEqual(["c", "b", "a"]);
    expect(sortTableRows(rows, null, "ascending")).toBe(rows);
  });
});

describe("useSortTableData", () => {
  let container: HTMLDivElement;

  const displayedNames = () =>
    Array.from(container.querySelectorAll("tbody tr")).map(
      (row) => row.querySelectorAll("td")[1]?.textContent,
    );

  const sortButton = () => container.querySelector(".p-table__sort-button");

  beforeAll(() => {
    (
      globalThis as unknown as { IS_REACT_ACT_ENVIRONMENT: boolean }
    ).IS_REACT_ACT_ENVIRONMENT = true;
  });

  beforeEach(() => {
    container = document.createElement("div");
    document.body.appendChild(container);
    const root = createRoot(container);
    act(() => {
      root.render(<Harness />);
    });
  });

  afterEach(() => {
    container.remove();
  });

  it("sorts all rows, not only the current page", () => {
    click(sortButton());
    expect(displayedNames()[0]).toBe("cluster-00");

    click(sortButton());
    expect(displayedNames()[0]).toBe("cluster-59");
  });

  it("selects the displayed range on shift-click after sorting", () => {
    click(sortButton());
    click(sortButton());

    const checkboxes = container.querySelectorAll<HTMLInputElement>(
      "tbody input[type='checkbox']",
    );
    click(checkboxes[0]);
    click(checkboxes[2], true);

    expect(container.querySelector("#selection")?.textContent).toBe(
      "cluster-59,cluster-58,cluster-57",
    );
  });
});

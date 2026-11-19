import { useId, useState, type DependencyList, type FC } from "react";
import {
  Icon,
  ScrollableContainer,
  ScrollableTable,
} from "@canonical/react-components";
import SelectableMainTable from "components/SelectableMainTable";
import PermissionGroupsFilter from "./PermissionGroupsFilter";
import { pluralize } from "util/helpers";
import useSortTableData from "util/useSortTableData";
import classnames from "classnames";
import type { LxdAuthGroup } from "types/permissions";

interface Props {
  groups: LxdAuthGroup[];
  modifiedGroups: Set<string>;
  parentItemName: string;
  parentItems?: { name: string }[];
  selectedGroups: Set<string>;
  setSelectedGroups: (val: string[], isUnselectAll?: boolean) => void;
  indeterminateGroups?: Set<string>;
  toggleGroup: (rowName: string) => void;
  scrollDependencies: DependencyList;
  preselectedGroups?: Set<string>;
  belowIds?: string[];
  disabled?: boolean;
  hasScrollableTable?: boolean;
  labelledBy?: string;
  searchLabelledBy?: string;
}

const GroupSelection: FC<Props> = ({
  groups,
  modifiedGroups,
  parentItemName,
  parentItems,
  selectedGroups,
  indeterminateGroups,
  setSelectedGroups,
  toggleGroup,
  scrollDependencies,
  preselectedGroups,
  belowIds = [],
  disabled = false,
  hasScrollableTable = true,
  labelledBy,
  searchLabelledBy,
}) => {
  const [search, setSearch] = useState("");
  const tableId = useId();

  const headers = [
    {
      content: "Group",
      sortKey: "name",
      className: "name",
    },
    {
      content: "Description",
      sortKey: "description",
      className: "description",
    },
    {
      content: "",
      "aria-label": "Modified status",
      className: "modified-status",
    },
  ];

  const filteredGroups = groups.filter((group) => {
    return group.name.toLowerCase().includes(search) || !search;
  });

  const rows = filteredGroups.map((group) => {
    const groupAdded =
      selectedGroups.has(group.name) && modifiedGroups.has(group.name);
    const groupRemoved =
      !selectedGroups.has(group.name) && modifiedGroups.has(group.name);

    const selectedParentsText =
      (parentItems?.length || 0) > 1
        ? `all selected ${pluralize(parentItemName, 2)}`
        : parentItems?.[0]?.name
          ? `${parentItemName} ${parentItems[0].name}`
          : parentItemName;
    const modifiedTitle = groupAdded
      ? `Group will be added to ${selectedParentsText}`
      : groupRemoved
        ? `Group will be removed from ${selectedParentsText}`
        : "";

    const toggleRow = () => {
      if (!disabled) {
        toggleGroup(group.name);
      }
    };

    return {
      key: group.name,
      name: group.name,
      className: "u-row",
      columns: [
        {
          content: group.name,
          title: group.name,
          onClick: toggleRow,
          role: "rowheader",
          className: classnames("name u-truncate", {
            "clickable-cell": !disabled,
          }),
          "aria-label": "Name",
        },
        {
          content: <span>{group.description || ""}</span>,
          onClick: toggleRow,
          role: "cell",
          className: classnames("description", {
            "clickable-cell": !disabled,
          }),
          "aria-label": "Description",
          title: group.description,
        },
        {
          content: modifiedGroups.has(group.name) && (
            <Icon name="status-in-progress-small" />
          ),
          role: "cell",
          "aria-label": "Modified status",
          className: "modified-status u-align--right",
          title: parentItemName ? modifiedTitle : undefined,
        },
      ],
      sortData: {
        name: group.name.toLowerCase(),
        description: group.description.toLowerCase(),
        isPreselected: preselectedGroups?.has(group.name),
      },
    };
  });

  const { rows: sortedRows } = useSortTableData({
    rows,
    defaultSort: preselectedGroups ? "isPreselected" : "name",
    defaultSortDirection: preselectedGroups ? "descending" : "ascending",
  });

  const filteredNames = filteredGroups.map((group) => group.name);
  const visibleNames = new Set(filteredNames);
  const updateVisibleSelection = (names: string[]) => {
    const selection = [
      ...[...selectedGroups].filter((name) => !visibleNames.has(name)),
      ...names,
    ];
    setSelectedGroups(selection, selection.length === 0);
  };

  const table = (
    <SelectableMainTable
      id={tableId}
      aria-labelledby={labelledBy}
      className="group-selection-table"
      headers={headers}
      rows={sortedRows}
      sortable
      emptyStateMsg={
        groups.length ? "No matching groups found" : "No groups found"
      }
      itemName="group"
      parentName=""
      selectedNames={[...selectedGroups].filter((name) =>
        visibleNames.has(name),
      )}
      setSelectedNames={updateVisibleSelection}
      disabledNames={[]}
      filteredNames={filteredNames}
      indeterminateNames={[...(indeterminateGroups ?? [])].filter((name) =>
        visibleNames.has(name),
      )}
      onToggleRow={toggleGroup}
      hideContextualMenu
      disableSelect={disabled}
      disableSelectAll={filteredNames.length === 0}
    />
  );

  return (
    <div className="group-selection">
      {groups.length > 5 && (
        <PermissionGroupsFilter
          onChange={setSearch}
          value={search}
          disabled={disabled}
          labelledBy={searchLabelledBy}
        />
      )}
      {hasScrollableTable ? (
        <ScrollableContainer
          dependencies={scrollDependencies}
          belowIds={["panel-footer", ...belowIds]}
        >
          <ScrollableTable
            dependencies={[...scrollDependencies, search]}
            tableId={tableId}
            belowIds={["panel-footer", ...belowIds]}
          >
            {table}
          </ScrollableTable>
        </ScrollableContainer>
      ) : (
        table
      )}
    </div>
  );
};

export default GroupSelection;

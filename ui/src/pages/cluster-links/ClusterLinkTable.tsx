import type { FC, ReactNode } from "react";
import { Fragment, useState } from "react";
import {
  Button,
  CheckboxInput,
  Icon,
  Spinner,
  Table,
  TableCell,
  TableHeader,
  TableRow,
} from "@canonical/react-components";
import { Link } from "react-router-dom";
import type {
  ClusterConnection,
  ClusterLinkSortKey,
  SourceClusterGroup,
} from "util/clusterLinks";
import { groupClusterLinks } from "util/clusterLinks";
import type { ClusterLinkSort } from "util/useSortClusterLinks";
import ClusterLinkRow from "./ClusterLinkRow";

interface Props extends ClusterLinkSort {
  rows?: ClusterConnection[];
  isLoading: boolean;
  processingIds: string[];
  selectedIds: string[];
  setSelectedIds: (ids: string[]) => void;
  groupBySource?: boolean;
}

const headers: {
  content: string;
  sortKey: ClusterLinkSortKey;
  className: string;
}[] = [
  { content: "Cluster", sortKey: "cluster", className: "cluster" },
  { content: "Link name", sortKey: "name", className: "name" },
  { content: "Link type", sortKey: "type", className: "type" },
  { content: "Target", sortKey: "target", className: "target" },
  {
    content: "Description",
    sortKey: "description",
    className: "description",
  },
  { content: "Used by", sortKey: "usedBy", className: "used-by" },
  { content: "Addresses", sortKey: "addresses", className: "addresses" },
];

const ClusterLinkTable: FC<Props> = ({
  rows = [],
  isLoading,
  processingIds,
  selectedIds,
  setSelectedIds,
  sortKey,
  sortDirection,
  updateSort,
  groupBySource = false,
}) => {
  const [collapsedGroups, setCollapsedGroups] = useState<string[]>([]);
  const groups = groupBySource ? groupClusterLinks(rows) : [];
  const activeHeaders = headers.filter(
    (header) => groupBySource || header.sortKey !== "cluster",
  );
  const columnCount = activeHeaders.length + 2;
  const pageIds = rows.map((row) => row.id);
  const selectedLookup = new Set(selectedIds);
  const processingLookup = new Set(processingIds);
  const isAllSelected =
    pageIds.length > 0 && pageIds.every((id) => selectedLookup.has(id));
  const hasSelection = selectedIds.length > 0;
  const isProcessing = processingIds.length > 0;

  const toggleIds = (ids: string[], select: boolean): void => {
    if (select) {
      setSelectedIds([...new Set([...selectedIds, ...ids])]);
    } else {
      const removed = new Set(ids);
      setSelectedIds(selectedIds.filter((id) => !removed.has(id)));
    }
  };

  const toggleGroup = (clusterName: string): void => {
    setCollapsedGroups((prev) =>
      prev.includes(clusterName)
        ? prev.filter((name) => name !== clusterName)
        : [...prev, clusterName],
    );
  };

  const renderRow = (entry: ClusterConnection): ReactNode => (
    <ClusterLinkRow
      key={entry.id}
      entry={entry}
      groupBySource={groupBySource}
      isSelected={selectedLookup.has(entry.id)}
      isDisabled={processingLookup.has(entry.id)}
      onSelect={() => {
        toggleIds([entry.id], !selectedLookup.has(entry.id));
      }}
    />
  );

  const renderGroup = (group: SourceClusterGroup): ReactNode => {
    const groupIds = group.links.map((entry) => entry.id);
    const isGroupSelected = groupIds.every((id) => selectedLookup.has(id));
    const isGroupPartiallySelected = groupIds.some((id) =>
      selectedLookup.has(id),
    );
    const isCollapsed = collapsedGroups.includes(group.cluster.name);

    return (
      <Fragment key={`group-${group.cluster.name}`}>
        <TableRow className="cluster-link-group-row">
          <TableCell className="select">
            <CheckboxInput
              label={
                <div className="u-off-screen">
                  Select links on {group.cluster.name}
                </div>
              }
              labelClassName="u-no-margin--bottom"
              checked={isGroupSelected}
              indeterminate={isGroupPartiallySelected && !isGroupSelected}
              disabled={isProcessing}
              onChange={() => {
                toggleIds(groupIds, !isGroupSelected);
              }}
            />
          </TableCell>
          <TableCell
            className="cluster cluster-link-group-name"
            title={group.cluster.name}
          >
            <Button
              appearance="base"
              dense
              className="u-no-margin--bottom"
              aria-label={`${
                isCollapsed ? "Expand" : "Collapse"
              } links on ${group.cluster.name}`}
              onClick={() => {
                toggleGroup(group.cluster.name);
              }}
            >
              <Icon name={isCollapsed ? "chevron-right" : "chevron-down"} />
            </Button>
            <Link to={`/ui/cluster/${group.cluster.name}`}>
              {group.cluster.name}
            </Link>
          </TableCell>
          <TableCell
            colSpan={columnCount - 2}
            className="cluster-link-group-spacer"
          />
        </TableRow>
        {!isCollapsed && group.links.map(renderRow)}
      </Fragment>
    );
  };

  const renderBody = (): ReactNode => {
    if (rows.length === 0) {
      return (
        <TableRow>
          <TableCell colSpan={columnCount}>
            {isLoading ? (
              <Spinner className="u-loader" text="Loading cluster links..." />
            ) : (
              "No cluster links found matching this search."
            )}
          </TableCell>
        </TableRow>
      );
    }

    return groupBySource ? groups.map(renderGroup) : rows.map(renderRow);
  };

  return (
    <>
      {isLoading && rows.length > 0 && (
        <p className="u-no-margin--bottom">
          <Spinner role="status" text="Loading cluster links..." />
        </p>
      )}
      <Table
        id="cluster-link-table"
        responsive
        className="cluster-link-table"
        aria-label="Cluster links"
      >
        <thead>
          <TableRow>
            <TableHeader className="select select-header" aria-label="select">
              <CheckboxInput
                label={<div className="u-off-screen">Select all</div>}
                labelClassName="multiselect-checkbox"
                className="multiselect-checkbox"
                checked={isAllSelected}
                indeterminate={hasSelection && !isAllSelected}
                disabled={isProcessing}
                onChange={() => {
                  setSelectedIds(hasSelection ? [] : pageIds);
                }}
              />
            </TableHeader>
            {activeHeaders.map((header) => (
              <TableHeader
                key={header.sortKey}
                className={header.className}
                title={header.content}
                sort={sortKey === header.sortKey ? sortDirection : "none"}
                onSort={() => {
                  updateSort(header.sortKey);
                }}
              >
                {header.content}
              </TableHeader>
            ))}
            <TableHeader
              className="u-align--right actions"
              aria-label="Actions"
            >
              <Icon name="settings" />
            </TableHeader>
          </TableRow>
        </thead>
        <tbody>{renderBody()}</tbody>
      </Table>
    </>
  );
};

export default ClusterLinkTable;

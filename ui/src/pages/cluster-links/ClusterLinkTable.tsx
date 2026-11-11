import type { FC } from "react";
import { useState } from "react";
import type {
  MainTableHeader,
  MainTableRow,
} from "@canonical/react-components/dist/components/MainTable/MainTable";
import { Button, Icon, MainTable, Spinner } from "@canonical/react-components";
import { Link } from "react-router-dom";
import type {
  ClusterConnection,
  GlobalClusterLinkSourceGroup,
} from "util/clusterLinks";
import {
  getGlobalClusterLinkGroupSortData,
  groupGlobalClusterLinks,
  sortGlobalClusterLinkRows,
} from "util/clusterLinks";
import getClusterLinkRow from "./ClusterLinkRow";

interface Props {
  rows?: ClusterConnection[];
  isLoading: boolean;
  groupBySource?: boolean;
  // Lets the parent sort all rows before they are paginated.
  onUpdateSort?: (sortKey?: string | null) => void;
}

const linkHeaders: MainTableHeader[] = [
  {
    content: "Link name",
    sortKey: "name",
    className: "name",
    title: "Link name",
  },
  {
    content: "Link type",
    sortKey: "type",
    className: "type",
    title: "Link type",
  },
  {
    content: "Target",
    sortKey: "target",
    className: "target",
    title: "Target",
  },
  {
    content: "Description",
    sortKey: "description",
    className: "description",
    title: "Description",
  },
  {
    content: "Used by",
    sortKey: "usedBy",
    className: "used-by",
    title: "Used by",
  },
  {
    content: "Addresses",
    sortKey: "addresses",
    className: "addresses",
    title: "Addresses",
  },
  {
    content: <Icon name="settings" />,
    "aria-label": "Actions",
    className: "u-align--right actions",
  },
];

const ClusterLinkTable: FC<Props> = ({
  rows = [],
  isLoading,
  groupBySource = false,
  onUpdateSort,
}) => {
  const [collapsedGroups, setCollapsedGroups] = useState<string[]>([]);
  const groups = groupBySource ? groupGlobalClusterLinks(rows) : [];

  const headers: MainTableHeader[] = groupBySource
    ? [
        {
          className: "cluster-link-toggle",
          "aria-label": "Expand or collapse cluster links",
        },
        {
          content: "Cluster",
          sortKey: "cluster",
          className: "cluster",
          title: "Cluster",
        },
        ...linkHeaders,
      ]
    : linkHeaders;

  const toggleGroup = (clusterName: string): void => {
    setCollapsedGroups((prev) =>
      prev.includes(clusterName)
        ? prev.filter((name) => name !== clusterName)
        : [...prev, clusterName],
    );
  };

  const getGroupRows = (
    group: GlobalClusterLinkSourceGroup,
    groupIndex: number,
  ): MainTableRow[] => {
    const isCollapsed = collapsedGroups.includes(group.cluster.name);
    const groupRow: MainTableRow = {
      key: `group-${group.cluster.name}`,
      className: "cluster-link-group-row",
      sortData: getGlobalClusterLinkGroupSortData(
        group.cluster.name,
        groupIndex,
      ),
      columns: [
        {
          className: "cluster-link-toggle",
          content: (
            <Button
              appearance="base"
              dense
              className="u-no-margin--bottom u-no-padding"
              aria-label={`${
                isCollapsed ? "Expand" : "Collapse"
              } links on ${group.cluster.name}`}
              aria-expanded={!isCollapsed}
              onClick={() => {
                toggleGroup(group.cluster.name);
              }}
            >
              <Icon name={isCollapsed ? "chevron-right" : "chevron-down"} />
            </Button>
          ),
        },
        {
          className: "cluster cluster-link-group-name",
          title: group.cluster.name,
          content: (
            <Link
              to={`/ui/cluster/${encodeURIComponent(group.cluster.name)}`}
              className="cluster-link-source-link"
            >
              {group.cluster.name}
            </Link>
          ),
        },
        { colSpan: headers.length - 2 },
      ],
    };

    return [
      groupRow,
      ...(isCollapsed
        ? []
        : group.links.map((entry) =>
            getClusterLinkRow(entry, groupBySource, groupIndex),
          )),
    ];
  };

  const tableRows: MainTableRow[] = groupBySource
    ? groups.flatMap(getGroupRows)
    : rows.map((entry) => getClusterLinkRow(entry, groupBySource, 0));

  return (
    <>
      {isLoading && rows.length > 0 && (
        <p className="u-no-margin--bottom">
          <Spinner role="status" text="Loading cluster links..." />
        </p>
      )}
      <MainTable
        id="cluster-link-table"
        className="cluster-link-table"
        aria-label="Cluster links"
        sortable
        sortFunction={groupBySource ? sortGlobalClusterLinkRows : undefined}
        onUpdateSort={onUpdateSort}
        headers={headers}
        rows={tableRows}
        emptyStateMsg={
          isLoading ? (
            <Spinner className="u-loader" text="Loading cluster links..." />
          ) : (
            "No cluster links found matching this search."
          )
        }
      />
    </>
  );
};

export default ClusterLinkTable;

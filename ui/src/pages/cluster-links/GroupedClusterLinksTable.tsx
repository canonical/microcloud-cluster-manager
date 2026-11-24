import type { FC } from "react";
import { useState } from "react";
import type {
  MainTableHeader,
  MainTableRow,
} from "@canonical/react-components/dist/components/MainTable/MainTable";
import { Button, Icon, MainTable, Spinner } from "@canonical/react-components";
import { Link } from "react-router-dom";
import type { ResolvedClusterLink } from "types/cluster";
import type { GlobalClusterLinkSourceGroup } from "util/clusterLinks";
import {
  CLUSTER_LINK_GROUP_SORT_COLUMN,
  getGlobalClusterLinkGroupSortData,
  getGlobalClusterLinkSortData,
  groupGlobalClusterLinks,
  sortGlobalClusterLinkRows,
} from "util/clusterLinks";
import getClusterLinkRow, {
  clusterLinkHeaders,
} from "components/cluster-links/ClusterLinkRow";

const groupColumns = [
  { className: "cluster-link-toggle" },
  { className: "cluster" },
];

interface Props {
  rows?: ResolvedClusterLink[];
  isLoading: boolean;
  onUpdateSort?: (sortKey?: string | null) => void;
}

const GroupedClusterLinksTable: FC<Props> = ({
  rows = [],
  isLoading,
  onUpdateSort,
}) => {
  const [collapsedGroups, setCollapsedGroups] = useState<string[]>([]);
  const groups = groupGlobalClusterLinks(rows);

  const headers: MainTableHeader[] = [
    {
      className: "cluster-link-toggle",
      "aria-label": "Expand or collapse cluster links",
    },
    {
      content: "Cluster",
      sortKey: CLUSTER_LINK_GROUP_SORT_COLUMN,
      className: "cluster",
      title: "Cluster",
    },
    ...clusterLinkHeaders,
  ];

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
    const isCollapsed = collapsedGroups.includes(group.sourceCluster);
    const groupRow: MainTableRow = {
      key: `group-${group.sourceCluster}`,
      className: "cluster-link-group-row",
      sortData: getGlobalClusterLinkGroupSortData(
        group.sourceCluster,
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
              } links on ${group.sourceCluster}`}
              aria-expanded={!isCollapsed}
              onClick={() => {
                toggleGroup(group.sourceCluster);
              }}
            >
              <Icon name={isCollapsed ? "chevron-right" : "chevron-down"} />
            </Button>
          ),
        },
        {
          className: "cluster cluster-link-group-name",
          title: group.sourceCluster,
          content: (
            <Link to={`/ui/cluster/${encodeURIComponent(group.sourceCluster)}`}>
              {group.sourceCluster}
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
        : group.links.map((resolvedLink) => {
            const row = getClusterLinkRow(resolvedLink);
            return {
              ...row,
              columns: [...groupColumns, ...row.columns],
              sortData: getGlobalClusterLinkSortData(resolvedLink, groupIndex),
            };
          })),
    ];
  };

  const tableRows = groups.flatMap(getGroupRows);

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
        sortFunction={sortGlobalClusterLinkRows}
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

export default GroupedClusterLinksTable;

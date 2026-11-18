import {
  Button,
  Icon,
  List,
  TablePagination,
  Tooltip,
} from "@canonical/react-components";
import type { FC } from "react";
import { useState } from "react";
import { Link } from "react-router-dom";
import useSortTableData from "context/useSortTableData";
import ClusterLinkTypeChip from "components/ClusterLinkTypeChip";
import DsIcon from "components/DsIcon";
import SelectedTableNotification from "components/SelectedTableNotification";
import SelectableMainTable from "components/SelectableMainTable";
import EditClusterLinkBtn from "pages/clusters/actions/EditClusterLinkBtn";
import DeleteClusterLinkBtn from "pages/clusters/actions/DeleteClusterLinkBtn";
import CreateClusterLinkBtn from "pages/clusters/actions/CreateClusterLinkBtn";
import type { Cluster, ClusterLink } from "types/cluster";
import {
  getLinkDisplayAddresses,
  getLinkedCluster,
  getLinkUsage,
} from "util/clusterLinks";
import { pluralize } from "util/helpers";

interface Props {
  cluster: Cluster;
  clusters: Cluster[];
  links: ClusterLink[];
}

const PAGE_SIZE = 10;

const ClusterLinksTable: FC<Props> = ({ cluster, clusters, links }) => {
  const [selectedNames, setSelectedNames] = useState<string[]>([]);
  const validSelectedNames = selectedNames.filter((name) =>
    links.some((link) => link.name === name),
  );

  const headers = [
    { content: "Link name", sortKey: "name", className: "name" },
    { content: "Link type", sortKey: "type", className: "type" },
    { content: "Target", sortKey: "target", className: "target" },
    {
      content: "Description",
      sortKey: "description",
      className: "description",
    },
    {
      content: (
        <span className="used-by-header">
          Used by
          <Tooltip
            position="btm-right"
            tooltipClassName="used-by-help-tooltip"
            message="Number of replicators and image registries using the cluster link. Hover over each cell for details."
          >
            <Icon name="help" className="used-by-help-icon" />
          </Tooltip>
        </span>
      ),
      sortKey: "usedBy",
      className: "used-by u-align--right",
    },
    { content: "Addresses", sortKey: "addresses", className: "addresses" },
    {
      content: <></>,
      "aria-label": "Actions",
      className: "u-align--right actions",
    },
  ];

  const rows = links.map((link) => {
    // The cluster this link points at, if it is known to the cluster manager.
    const targetCluster = getLinkedCluster(link, clusters);
    const shownAddresses = getLinkDisplayAddresses(link);
    const usage = getLinkUsage(link);
    const usedByCount = link.used_by?.length ?? 0;

    return {
      key: link.name,
      name: link.name,
      columns: [
        { content: link.name, role: "rowheader", className: "name" },
        {
          content: <ClusterLinkTypeChip type={link.type} />,
          role: "cell",
          className: "type",
        },
        {
          content: targetCluster ? (
            <Link to={`/ui/cluster/${encodeURIComponent(targetCluster.name)}`}>
              {targetCluster.name}
            </Link>
          ) : (
            "–"
          ),
          role: "cell",
          className: "target",
        },
        {
          content: link.description ? (
            <div className="description-text">{link.description}</div>
          ) : (
            "–"
          ),
          role: "cell",
          className: "description",
          title: link.description,
        },
        {
          content: usedByCount ? (
            <Tooltip
              position="btm-left"
              tooltipClassName="cell-details-tooltip"
              message={
                <>
                  {usage.replicators > 0 && (
                    <div>
                      {usage.replicators}{" "}
                      {pluralize("replicator", usage.replicators)}
                    </div>
                  )}
                  {usage.imageRegistries > 0 && (
                    <div>
                      {usage.imageRegistries}{" "}
                      {usage.imageRegistries === 1
                        ? "image registry"
                        : "image registries"}
                    </div>
                  )}
                </>
              }
            >
              <span tabIndex={0}>{usedByCount}</span>
            </Tooltip>
          ) : (
            usedByCount
          ),
          role: "cell",
          className: "used-by u-align--right",
        },
        {
          content: shownAddresses.length ? (
            <>
              {shownAddresses
                .slice(0, shownAddresses.length > 2 ? 1 : 2)
                .map((address, index) => (
                  <div key={`${address}-${index}`}>{address}</div>
                ))}
              {shownAddresses.length > 2 && (
                <Tooltip
                  position="btm-left"
                  tooltipClassName="cell-details-tooltip"
                  message={shownAddresses.map((address, index) => (
                    <div key={`${address}-${index}`}>{address}</div>
                  ))}
                >
                  <span className="u-text--muted" tabIndex={0}>
                    +{shownAddresses.length - 1}
                  </span>
                </Tooltip>
              )}
            </>
          ) : (
            "–"
          ),
          role: "cell",
          className: "addresses",
        },
        {
          content: (
            <List
              inline
              className="u-no-margin--bottom actions-list"
              items={[
                <EditClusterLinkBtn key="edit" cluster={cluster} link={link} />,
                <DeleteClusterLinkBtn
                  key="delete"
                  cluster={cluster}
                  link={link}
                  targetCluster={targetCluster}
                  iconOnly
                />,
              ]}
            />
          ),
          role: "cell",
          className: "u-align--right actions",
        },
      ],
      sortData: {
        name: link.name.toLowerCase(),
        type: link.type,
        target: (targetCluster?.name ?? "").toLowerCase(),
        description: (link.description ?? "").toLowerCase(),
        usedBy: usedByCount,
        addresses: (shownAddresses[0] ?? "").toLowerCase(),
      },
    };
  });

  const { rows: sortedRows, updateSort } = useSortTableData({ rows });

  const selectedNotification = validSelectedNames.length > 0 && (
    <SelectedTableNotification
      totalCount={links.length}
      itemName="link"
      parentName="cluster"
      selectedNames={validSelectedNames}
      setSelectedNames={setSelectedNames}
      filteredNames={links.map((link) => link.name)}
    />
  );

  const table = (
    <SelectableMainTable
      id="cluster-links-table"
      className="cluster-links-table"
      responsive
      expanding={false}
      sortable
      headers={headers}
      rows={sortedRows}
      onUpdateSort={updateSort}
      selectedNames={validSelectedNames}
      setSelectedNames={setSelectedNames}
      itemName="link"
      parentName="cluster"
      filteredNames={links.map((link) => link.name)}
      disabledNames={[]}
      hideContextualMenu
    />
  );

  return (
    <>
      <div className="cluster-links-create">
        <CreateClusterLinkBtn cluster={cluster} />
      </div>
      <div className="cluster-links-header">
        <h2 className="p-heading--4 u-no-margin--bottom u-no-padding--top">
          Cluster link details
        </h2>
        {validSelectedNames.length > 1 && (
          // Deleting several links at once is not supported yet
          <Button className="u-no-margin--bottom has-icon" disabled>
            <DsIcon icon="delete" />
            <span>Delete</span>
          </Button>
        )}
      </div>
      {links.length > PAGE_SIZE ? (
        <TablePagination
          data={sortedRows}
          id="cluster-links-pagination"
          itemName="link"
          className="u-no-margin--top u-no-margin--bottom"
          aria-label="Table pagination control"
          pageLimits={[PAGE_SIZE, 25, 50]}
          position="below"
          description={selectedNotification}
        >
          {table}
        </TablePagination>
      ) : (
        <>
          {table}
          {selectedNotification && (
            <div className="cluster-links-selected">{selectedNotification}</div>
          )}
        </>
      )}
    </>
  );
};

export default ClusterLinksTable;

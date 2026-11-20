import { Icon, Tooltip } from "@canonical/react-components";
import type {
  MainTableHeader,
  MainTableRow,
} from "@canonical/react-components/dist/components/MainTable/MainTable";
import { Link } from "react-router-dom";
import type { ClusterConnection } from "util/clusterLinks";
import { getClusterLinkSortData } from "util/clusterLinks";
import ClusterLinkTypeChip from "components/cluster-links/ClusterLinkTypeChip";
import ActionsList from "components/ActionsList";
import EditClusterLinkButton from "components/cluster-links/EditClusterLinkButton";
import DeleteClusterLinksButton from "components/cluster-links/DeleteClusterLinksButton";
import ClusterLinkAddressesCell from "./ClusterLinkAddressesCell";
import ClusterLinkUsedByCell from "./ClusterLinkUsedByCell";

export const clusterLinkHeaders: MainTableHeader[] = [
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
    content: (
      <>
        Used by{" "}
        <Tooltip message="Number of replicators, image registries, and projects using the cluster link. Open the count to see details.">
          <Icon
            name="help"
            className="u-text--muted cluster-link-used-by-help"
          />
        </Tooltip>
      </>
    ),
    sortKey: "usedBy",
    className: "used-by u-align--right",
    title: "Used by",
  },
  {
    content: "Addresses",
    sortKey: "addresses",
    className: "addresses",
    title: "Addresses",
  },
  {
    content: "",
    "aria-label": "Actions",
    className: "u-align--right actions",
  },
];

const getClusterLinkRow = (
  entry: ClusterConnection,
): MainTableRow & {
  columns: NonNullable<MainTableRow["columns"]>;
} => {
  return {
    key: entry.id,
    name: entry.link.name,
    className: "cluster-link-row",
    sortData: getClusterLinkSortData(entry),
    columns: [
      {
        role: "rowheader",
        className: "name",
        title: entry.link.name,
        content: (
          <>
            <Link
              to={`/ui/cluster/${encodeURIComponent(entry.source.name)}/cluster-links`}
              className="u-truncate cluster-link-name-link"
            >
              {entry.link.name}
            </Link>
          </>
        ),
      },
      {
        className: "type",
        title: entry.link.type,
        "aria-label": "Link type",
        content: <ClusterLinkTypeChip type={entry.link.type} />,
      },
      {
        className: "target",
        title: entry.target?.name,
        content: (
          <>
            {entry.target ? (
              <Link to={`/ui/cluster/${encodeURIComponent(entry.target.name)}`}>
                {entry.target.name}
              </Link>
            ) : (
              <>
                unknown
                <Tooltip
                  message="The linked cluster might not be enrolled in this cluster manager."
                  position="right"
                >
                  <Icon
                    name="help"
                    className="u-text--muted cluster-link-unknown-target-icon"
                  />
                </Tooltip>
              </>
            )}
          </>
        ),
      },
      {
        className: "description",
        title: entry.link.description,
        content: entry.link.description ? (
          <div className="cluster-link-description">
            {entry.link.description}
          </div>
        ) : (
          "–"
        ),
      },
      {
        className: "used-by u-align--right",
        content: <ClusterLinkUsedByCell link={entry.link} />,
      },
      {
        className: "addresses",
        content: <ClusterLinkAddressesCell entry={entry} />,
      },
      {
        className: "u-align--right actions",
        hasOverflow: true,
        content: (
          <ActionsList
            items={[
              <EditClusterLinkButton
                cluster={entry.source}
                link={entry.link}
                hasLabel={false}
                key="edit"
              />,
              <DeleteClusterLinksButton
                cluster={entry.source.name}
                links={[entry]}
                hasLabel={false}
                key="delete"
              />,
            ]}
          />
        ),
      },
    ],
  };
};

export default getClusterLinkRow;

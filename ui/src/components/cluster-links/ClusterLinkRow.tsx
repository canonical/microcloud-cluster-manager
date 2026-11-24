import { Icon, Tooltip } from "@canonical/react-components";
import type {
  MainTableHeader,
  MainTableRow,
} from "@canonical/react-components/dist/components/MainTable/MainTable";
import { Link } from "react-router-dom";
import type { ClusterConnection } from "util/clusterLinks";
import { getClusterLinkSortData } from "util/clusterLinks";
import ClusterLinkTypeChip from "components/cluster-links/ClusterLinkTypeChip";
import ClusterRichChip from "pages/clusters/ClusterRichChip";
import ActionsList from "components/ActionsList";
import EditClusterLinkBtn from "components/cluster-links/EditClusterLinkBtn";
import DeleteClusterLinkBtn from "components/cluster-links/DeleteClusterLinkBtn";
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
              className="u-truncate"
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
        content: (
          <>
            {entry.target ? (
              <ClusterRichChip
                clusterName={entry.target.name}
                hasTitle={false}
              />
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
              <EditClusterLinkBtn
                cluster={entry.source}
                link={entry.link}
                hasLabel={false}
                key="edit"
              />,
              <DeleteClusterLinkBtn entry={entry} key="delete" />,
            ]}
          />
        ),
      },
    ],
  };
};

export default getClusterLinkRow;

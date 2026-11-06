import { Chip, Icon, Tooltip } from "@canonical/react-components";
import type { MainTableRow } from "@canonical/react-components/dist/components/MainTable/MainTable";
import { Link } from "react-router-dom";
import type { ClusterConnection } from "util/clusterLinks";
import { getClusterLinkSortData, linkTypeIcons } from "util/clusterLinks";
import { capitalizeFirstLetter } from "util/helpers";
import ActionsList from "components/ActionsList";
import DeleteClusterLinkBtn from "pages/clusters/actions/DeleteClusterLinkBtn";
import ClusterLinkAddressesCell from "./ClusterLinkAddressesCell";

// Builds a MainTable row for a single cluster link. The sort data keeps link
// rows attached to their group and scoped to it when a column is sorted.
const getClusterLinkRow = (
  entry: ClusterConnection,
  groupBySource: boolean,
  group: number,
): MainTableRow => {
  return {
    key: entry.id,
    className: "cluster-link-row",
    sortData: getClusterLinkSortData(entry, group),
    columns: [
      ...(groupBySource
        ? [{ className: "cluster-link-toggle" }, { className: "cluster" }]
        : []),
      {
        role: "rowheader",
        className: "name",
        title: entry.link.name,
        content: (
          <>
            <span className="cluster-link-field-label">Link name</span>
            <Link
              to={`/ui/cluster/${entry.source.name}/cluster-links`}
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
        content: (
          <>
            <span className="cluster-link-field-label">Link type</span>
            <Chip
              isReadOnly
              isDense
              isInline
              iconName={linkTypeIcons[entry.link.type.toLowerCase()]}
              value={capitalizeFirstLetter(entry.link.type)}
            />
          </>
        ),
      },
      {
        className: "target",
        title: entry.target?.name,
        content: (
          <>
            <span className="cluster-link-field-label">Target cluster</span>
            {entry.target ? (
              <Link to={`/ui/cluster/${entry.target.name}`}>
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
        className: "description u-truncate",
        title: entry.link.description,
        content: entry.link.description,
      },
      {
        className: "used-by",
        content: entry.link.used_by?.length ?? 0,
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
              <DeleteClusterLinkBtn
                cluster={entry.source}
                link={entry.link}
                targetCluster={entry.target}
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

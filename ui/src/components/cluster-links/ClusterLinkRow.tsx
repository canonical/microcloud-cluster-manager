import { Icon, Tooltip } from "@canonical/react-components";
import type {
  MainTableHeader,
  MainTableRow,
} from "@canonical/react-components/dist/components/MainTable/MainTable";
import { Link } from "react-router-dom";
import type { ResolvedClusterLink } from "types/cluster";
import { getClusterLinkId, getClusterLinkSortData } from "util/clusterLinks";
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
  resolvedLink: ResolvedClusterLink,
): MainTableRow & {
  columns: NonNullable<MainTableRow["columns"]>;
} => {
  return {
    key: getClusterLinkId(resolvedLink),
    name: resolvedLink.link.name,
    className: "cluster-link-row",
    sortData: getClusterLinkSortData(resolvedLink),
    columns: [
      {
        role: "rowheader",
        className: "name",
        title: resolvedLink.link.name,
        content: (
          <>
            <Link
              to={`/ui/cluster/${encodeURIComponent(resolvedLink.source_cluster)}/cluster-links`}
              className="u-truncate cluster-link-name-link"
            >
              {resolvedLink.link.name}
            </Link>
          </>
        ),
      },
      {
        className: "type",
        title: resolvedLink.link.type,
        "aria-label": "Link type",
        content: <ClusterLinkTypeChip type={resolvedLink.link.type} />,
      },
      {
        className: "target",
        title: resolvedLink.target_cluster ?? undefined,
        content: (
          <>
            {resolvedLink.target_cluster ? (
              <Link
                to={`/ui/cluster/${encodeURIComponent(resolvedLink.target_cluster)}`}
              >
                {resolvedLink.target_cluster}
              </Link>
            ) : (
              <>
                unknown
                <Tooltip
                  message="The target cluster might not be managed by this cluster manager."
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
        title: resolvedLink.link.description,
        content: resolvedLink.link.description ? (
          <div className="cluster-link-description">
            {resolvedLink.link.description}
          </div>
        ) : (
          "–"
        ),
      },
      {
        className: "used-by u-align--right",
        content: <ClusterLinkUsedByCell link={resolvedLink.link} />,
      },
      {
        className: "addresses",
        content: <ClusterLinkAddressesCell resolvedLink={resolvedLink} />,
      },
      {
        className: "u-align--right actions",
        hasOverflow: true,
        content: (
          <ActionsList
            items={[
              <EditClusterLinkButton
                cluster={resolvedLink.source_cluster}
                link={resolvedLink.link}
                hasLabel={false}
                key="edit"
              />,
              <DeleteClusterLinksButton
                cluster={resolvedLink.source_cluster}
                links={[resolvedLink]}
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

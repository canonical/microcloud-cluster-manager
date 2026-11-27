import { Icon, Tooltip } from "@canonical/react-components";
import type {
  MainTableHeader,
  MainTableRow,
} from "@canonical/react-components/dist/components/MainTable/MainTable";
import { Link } from "react-router-dom";
import type { ResolvedClusterLink } from "types/cluster";
import { getClusterLinkId, getClusterLinkSortData } from "util/clusterLinks";
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
  resolvedLink: ResolvedClusterLink,
): MainTableRow & {
  columns: NonNullable<MainTableRow["columns"]>;
} => {
  const id = getClusterLinkId(resolvedLink);
  return {
    key: id,
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
              className="u-truncate"
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
        content: (
          <>
            {resolvedLink.target_cluster ? (
              <ClusterRichChip
                clusterName={resolvedLink.target_cluster}
                hasTitle={false}
              />
            ) : (
              <>
                unknown
                <Tooltip
                  message="The target cluster could not be identified."
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
              <EditClusterLinkBtn
                cluster={resolvedLink.source_cluster}
                link={resolvedLink.link}
                hasLabel={false}
                key="edit"
              />,
              <DeleteClusterLinkBtn resolvedLink={resolvedLink} key="delete" />,
            ]}
          />
        ),
      },
    ],
  };
};

export default getClusterLinkRow;

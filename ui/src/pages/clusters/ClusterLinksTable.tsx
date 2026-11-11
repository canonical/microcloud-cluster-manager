import type { FC } from "react";
import { useState } from "react";
import { Link } from "react-router-dom";
import DsIcon from "components/DsIcon";
import ClusterLinkTypeChip from "components/ClusterLinkTypeChip";
import SelectableMainTable from "components/SelectableMainTable";
import EditClusterLinkBtn from "pages/clusters/actions/EditClusterLinkBtn";
import DeleteClusterLinkBtn from "pages/clusters/actions/DeleteClusterLinkBtn";
import DeleteClusterLinksBtn from "pages/clusters/actions/DeleteClusterLinksBtn";
import CreateClusterLinkBtn from "pages/clusters/actions/CreateClusterLinkBtn";
import type { Cluster, ClusterLink } from "types/cluster";
import { getLinkAddresses, getLinkedCluster } from "util/clusterLinks";

interface Props {
  cluster: Cluster;
  clusters: Cluster[];
  links: ClusterLink[];
}

const ClusterLinksTable: FC<Props> = ({ cluster, clusters, links }) => {
  const [selectedNames, setSelectedNames] = useState<string[]>([]);
  const validSelectedNames = selectedNames.filter((name) =>
    links.some((link) => link.name === name),
  );

  const headers = [
    { content: "Link name", className: "name" },
    { content: "Link type", className: "type" },
    { content: "Target", className: "target" },
    { content: "Description", className: "description" },
    { content: "Used by", className: "used-by u-align--right" },
    { content: "Addresses", className: "addresses" },
    {
      content: (
        <DsIcon
          icon="settings"
          className="cluster-links-table__settings-icon"
        />
      ),
      "aria-label": "Actions",
      className: "u-align--right actions",
    },
  ];

  const rows = links.map((link) => {
    const targetCluster = getLinkedCluster(link, clusters);
    const addresses = getLinkAddresses(link);
    const pendingAddress = link.config?.["volatile.pending_address"];
    const shownAddresses = addresses.length
      ? addresses
      : pendingAddress
        ? [pendingAddress]
        : [];

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
            <Link to={`/ui/cluster/${targetCluster.name}`}>
              {targetCluster.name}
            </Link>
          ) : (
            "–"
          ),
          role: "cell",
          className: "target",
        },
        {
          content: link.description || "–",
          role: "cell",
          className: "description",
        },
        {
          content: link.used_by?.length || "–",
          role: "cell",
          className: "used-by u-align--right",
        },
        {
          content: shownAddresses.length ? (
            <>
              {shownAddresses.slice(0, 2).map((address) => (
                <div key={address}>{address}</div>
              ))}
              {shownAddresses.length > 2 && (
                <div className="u-text--muted">
                  +{shownAddresses.length - 2}
                </div>
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
            <>
              <EditClusterLinkBtn cluster={cluster} link={link} />
              <span className="cluster-links-table__divider" aria-hidden />
              <DeleteClusterLinkBtn
                cluster={cluster}
                link={link}
                targetCluster={targetCluster}
                iconOnly
              />
            </>
          ),
          role: "cell",
          className: "u-align--right actions",
        },
      ],
    };
  });

  const selectedLinks = links.filter((link) =>
    validSelectedNames.includes(link.name),
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
        {selectedLinks.length > 0 && (
          <DeleteClusterLinksBtn
            cluster={cluster}
            clusters={clusters}
            links={selectedLinks}
            onDeleted={() => {
              setSelectedNames([]);
            }}
          />
        )}
      </div>
      <SelectableMainTable
        id="cluster-links-table"
        className="cluster-links-table"
        responsive
        expanding={false}
        headers={headers}
        rows={rows}
        selectedNames={validSelectedNames}
        setSelectedNames={setSelectedNames}
        itemName="link"
        parentName="cluster"
        filteredNames={links.map((link) => link.name)}
        disabledNames={[]}
        hideContextualMenu
      />
    </>
  );
};

export default ClusterLinksTable;

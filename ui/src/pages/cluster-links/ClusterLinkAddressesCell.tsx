import type { FC } from "react";
import { useState } from "react";
import { ContextualMenu } from "@canonical/react-components";
import { Link } from "react-router-dom";
import type { ClusterConnection } from "util/clusterLinks";
import { getLinkAddresses } from "util/clusterLinks";

interface Props {
  entry: ClusterConnection;
}

const ClusterLinkAddressesCell: FC<Props> = ({ entry }) => {
  const addresses = getLinkAddresses(entry.link);
  const [showAddresses, setShowAddresses] = useState(false);

  return (
    <>
      <span className="cluster-link-field-label">Addresses</span>
      {addresses.length === 0 ? (
        "unknown"
      ) : (
        <>
          <span className="u-truncate" title={addresses[0]}>
            {addresses[0]}
          </span>
          {addresses.length === 2 && (
            <span className="u-truncate" title={addresses[1]}>
              {addresses[1]}
            </span>
          )}
          {addresses.length > 2 && (
            <ContextualMenu
              visible={showAddresses}
              onToggleMenu={setShowAddresses}
              onPointerEnter={(event) => {
                if (event.pointerType === "mouse") {
                  setShowAddresses(true);
                }
              }}
              onPointerLeave={(event) => {
                if (event.pointerType === "mouse") {
                  setShowAddresses(false);
                }
              }}
              position="right"
              toggle={
                <Link
                  to={`/ui/cluster/${entry.source.name}/cluster-links`}
                  className="cluster-link-addresses-more"
                  aria-label={`View cluster links on ${entry.source.name}`}
                >
                  +{addresses.length - 1}
                </Link>
              }
              dropdownClassName="cluster-link-addresses-menu"
              dropdownProps={{
                "aria-label": `Addresses for ${entry.link.name} on ${entry.source.name}`,
              }}
            >
              <ul className="p-list u-no-margin--bottom">
                {addresses.map((address) => (
                  <li key={address}>{address}</li>
                ))}
              </ul>
            </ContextualMenu>
          )}
        </>
      )}
    </>
  );
};

export default ClusterLinkAddressesCell;

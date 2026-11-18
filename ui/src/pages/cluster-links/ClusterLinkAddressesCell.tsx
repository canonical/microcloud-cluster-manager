import type { FC } from "react";
import { useState } from "react";
import { Button, ContextualMenu } from "@canonical/react-components";
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
                <Button
                  appearance="link"
                  className="cluster-link-addresses-more u-no-margin--bottom"
                  aria-label={`View cluster links on ${entry.source.name}`}
                >
                  +{addresses.length - 1}
                </Button>
              }
              dropdownClassName="cluster-link-addresses-menu"
              dropdownProps={{
                "aria-label": `Addresses for ${entry.link.name} on ${entry.source.name}`,
              }}
            >
              <ul className="p-list u-no-margin--bottom">
                {addresses.map((address, index) => (
                  <li key={`${address}-${index}`}>{address}</li>
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

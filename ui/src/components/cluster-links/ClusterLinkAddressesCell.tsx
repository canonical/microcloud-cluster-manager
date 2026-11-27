import type { FC } from "react";
import { ContextualMenu } from "@canonical/react-components";
import type { ResolvedClusterLink } from "types/cluster";
import { getLinkDisplayAddresses } from "util/clusterLinks";

interface Props {
  resolvedLink: ResolvedClusterLink;
}

const ClusterLinkAddressesCell: FC<Props> = ({ resolvedLink }) => {
  const addresses = getLinkDisplayAddresses(resolvedLink.link);
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
              position="right"
              toggleLabel={`+${addresses.length - 1}`}
              toggleAppearance="link"
              toggleClassName="u-no-margin--bottom"
              toggleProps={{
                "aria-label": `View more addresses for ${resolvedLink.link.name}`,
              }}
              dropdownClassName="cluster-link-addresses-menu"
              dropdownProps={{
                "aria-label": `Addresses for ${resolvedLink.link.name} on ${resolvedLink.source_cluster}`,
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

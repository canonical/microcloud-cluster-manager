import { Button, ContextualMenu } from "@canonical/react-components";
import type { FC } from "react";
import { useState } from "react";
import type { ClusterLink } from "types/cluster";
import ClusterLinkUsedBy from "./ClusterLinkUsedBy";

interface Props {
  link: ClusterLink;
}

const ClusterLinkUsedByCell: FC<Props> = ({ link }) => {
  const [showUsedBy, setShowUsedBy] = useState(false);
  const usedBy = link.used_by ?? [];

  if (usedBy.length === 0) {
    return <>0</>;
  }

  return (
    <ContextualMenu
      visible={showUsedBy}
      onToggleMenu={setShowUsedBy}
      onMouseEnter={() => {
        setShowUsedBy(true);
      }}
      onMouseLeave={() => {
        setShowUsedBy(false);
      }}
      position="right"
      // A custom toggle keeps the menu from focusing the number when hover ends.
      // Keyboard focus stays on the button because the details are non-interactive.
      toggle={
        <Button
          appearance="base"
          dense
          className="cluster-link-used-by-count u-no-margin--bottom"
          aria-label={`${usedBy.length} resources using ${link.name}`}
          aria-haspopup="menu"
          aria-expanded={showUsedBy}
          onFocus={() => {
            setShowUsedBy(true);
          }}
          onBlur={() => {
            setShowUsedBy(false);
          }}
          onClick={() => {
            setShowUsedBy(true);
          }}
        >
          {usedBy.length}
        </Button>
      }
      dropdownClassName="cluster-link-used-by-menu"
      dropdownProps={{
        "aria-label": `Usage details for ${link.name}`,
      }}
    >
      <ClusterLinkUsedBy usedBy={usedBy} />
    </ContextualMenu>
  );
};

export default ClusterLinkUsedByCell;

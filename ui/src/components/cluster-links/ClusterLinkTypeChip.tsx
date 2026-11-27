import { Tooltip } from "@canonical/react-components";
import type { FC } from "react";
import ResourceIcon from "components/ResourceIcon";
import { getLinkIconType, LINK_TYPES } from "util/clusterLinkTypes";
import type { ClusterLink } from "types/cluster";

interface Props {
  type: ClusterLink["type"];
}

const ClusterLinkTypeChip: FC<Props> = ({ type }) => {
  const linkType = LINK_TYPES.find((t) => t.value === type);
  const label = linkType?.label ?? type;

  const chip = (
    <span className="p-chip u-no-margin--bottom cluster-link-type-chip is-dense is-inline">
      <ResourceIcon
        type={getLinkIconType(type)}
        className="cluster-link-type-chip__icon"
      />
      <span className="p-chip__value">{label}</span>
    </span>
  );

  if (!linkType) {
    return chip;
  }

  return (
    <Tooltip
      message={linkType.description}
      position="right"
      tooltipClassName="cluster-link-type-tooltip"
    >
      {chip}
    </Tooltip>
  );
};

export default ClusterLinkTypeChip;

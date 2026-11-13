import type { FC } from "react";
import ResourceIcon from "components/ResourceIcon";
import { getLinkIconType, LINK_TYPES } from "util/clusterLinkTypes";
import type { ClusterLink } from "types/cluster";

interface Props {
  type: ClusterLink["type"];
}

const ClusterLinkTypeChip: FC<Props> = ({ type }) => {
  const label = LINK_TYPES.find((t) => t.value === type)?.label ?? type;

  return (
    <span className="p-chip is-readonly u-no-margin--bottom cluster-link-type-chip">
      <ResourceIcon
        type={getLinkIconType(type)}
        className="cluster-link-type-chip__icon"
      />
      <span className="p-chip__value">{label}</span>
    </span>
  );
};

export default ClusterLinkTypeChip;

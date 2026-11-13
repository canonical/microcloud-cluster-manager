import type { FC } from "react";
import { Tooltip } from "@canonical/react-components";
import { useIsScreenBelow } from "context/useIsScreenBelow";
import ResourceLink from "components/ResourceLink";
import { SMALL_TOOLTIP_BREAKPOINT } from "components/RichTooltipTable";
import ClusterLinkRichTooltip from "./ClusterLinkRichTooltip";
import type { ClusterLinkType } from "types/cluster";

interface Props {
  clusterLinkName: string;
  clusterLinkType: ClusterLinkType;
  cluster: string;
  className?: string;
  disabled?: boolean;
}

const ClusterLinkRichChip: FC<Props> = ({
  clusterLinkName,
  clusterLinkType,
  cluster,
  className,
  disabled,
}) => {
  const showTooltip = !useIsScreenBelow(SMALL_TOOLTIP_BREAKPOINT, "height");

  const resourceLink = (
    <ResourceLink
      type={`cluster-link-${clusterLinkType}`}
      value={clusterLinkName}
      hasTitle={!showTooltip}
      className={className}
      disabled={disabled}
    />
  );

  if (!showTooltip) {
    return <>{resourceLink}</>;
  }

  return (
    <Tooltip
      zIndex={1000}
      position="right"
      message={
        <ClusterLinkRichTooltip
          cluster={cluster}
          clusterLinkName={clusterLinkName}
        />
      }
    >
      {resourceLink}
    </Tooltip>
  );
};

export default ClusterLinkRichChip;

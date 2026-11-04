import type { FC } from "react";
import { Tooltip } from "@canonical/react-components";
import { useIsScreenBelow } from "context/useIsScreenBelow";
import ResourceLink from "components/ResourceLink";
import { SMALL_TOOLTIP_BREAKPOINT } from "components/RichTooltipTable";
import ClusterRichTooltip from "./ClusterRichTooltip";

interface Props {
  clusterName: string;
  className?: string;
  disabled?: boolean;
  urlSuffix?: string;
}

const ClusterRichChip: FC<Props> = ({
  clusterName,
  className,
  disabled,
  urlSuffix = "",
}) => {
  const showTooltip = !useIsScreenBelow(SMALL_TOOLTIP_BREAKPOINT, "height");

  const resourceLink = (
    <ResourceLink
      type="cluster"
      value={clusterName}
      to={`/ui/cluster/${clusterName}${urlSuffix}`}
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
      message={<ClusterRichTooltip clusterName={clusterName} />}
    >
      {resourceLink}
    </Tooltip>
  );
};

export default ClusterRichChip;

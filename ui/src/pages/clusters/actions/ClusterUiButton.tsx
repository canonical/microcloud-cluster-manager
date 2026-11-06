import type { FC } from "react";
import ExternalLinkButton from "components/ExternalLinkButton";

interface Props {
  lxdUrl: string;
  className?: string;
  onClose?: () => void;
  hasLabel?: boolean;
}

const ClusterUiButton: FC<Props> = ({
  lxdUrl,
  className,
  onClose,
  hasLabel = true,
}) => {
  return (
    <ExternalLinkButton
      href={lxdUrl || undefined}
      label="LXD UI"
      iconName="external-link"
      unavailableReason="The cluster has not reported an LXD UI address"
      className={className}
      onClose={onClose}
      hasLabel={hasLabel}
    />
  );
};

export default ClusterUiButton;

import { Button } from "@canonical/react-components";
import type { FC } from "react";
import classnames from "classnames";
import DsIcon from "components/DsIcon";

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
  const isAvailable = !!lxdUrl;
  let title: string | undefined;

  if (!isAvailable) {
    title = "The cluster has not reported an LXD UI address";
  } else if (!hasLabel) {
    title = "LXD UI";
  }

  return (
    <Button
      element={isAvailable ? "a" : "button"}
      appearance={hasLabel ? undefined : "base"}
      hasIcon
      className={classnames("u-no-margin--bottom", className)}
      disabled={!isAvailable}
      href={lxdUrl}
      target={isAvailable ? "_blank" : undefined}
      rel={isAvailable ? "noopener noreferrer" : undefined}
      onClick={isAvailable ? onClose : undefined}
      aria-label={
        hasLabel ? undefined : `LXD UI${isAvailable ? "" : " (unavailable)"}`
      }
      title={title}
    >
      <DsIcon icon="external-link" />
      {hasLabel && <span>LXD UI</span>}
    </Button>
  );
};

export default ClusterUiButton;

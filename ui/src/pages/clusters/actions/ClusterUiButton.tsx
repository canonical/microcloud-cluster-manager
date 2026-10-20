import { Icon } from "@canonical/react-components";
import type { FC } from "react";
import classnames from "classnames";

interface Props {
  lxdUrl: string;
  className?: string;
  onClose?: () => void;
}

const ClusterUiButton: FC<Props> = ({ lxdUrl, className, onClose }) => {
  if (!lxdUrl) {
    return null;
  }

  return (
    <a
      className={classnames("p-button u-no-margin--bottom has-icon", className)}
      onClick={onClose}
      href={lxdUrl}
      target="_blank"
      rel="noopener noreferrer"
    >
      <Icon name="external-link" />
      <span>LXD UI</span>
    </a>
  );
};

export default ClusterUiButton;

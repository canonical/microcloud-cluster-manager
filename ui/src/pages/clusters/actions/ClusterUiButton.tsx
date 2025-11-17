import { Icon } from "@canonical/react-components";
import type { FC } from "react";
import classnames from "classnames";

interface Props {
  uiUrl: string;
  className?: string;
  onClose?: () => void;
}

const ClusterUiButton: FC<Props> = ({ uiUrl, className, onClose }) => {
  if (!uiUrl) {
    return null;
  }

  return (
    <a
      className={classnames("p-button u-no-margin--bottom has-icon", className)}
      onClick={onClose}
      href={uiUrl}
      target="_blank"
      rel="noopener noreferrer"
    >
      <Icon name="external-link" />
      <span>LXD UI</span>
    </a>
  );
};

export default ClusterUiButton;

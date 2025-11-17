import { ActionButton, Icon } from "@canonical/react-components";
import type { FC } from "react";
import classnames from "classnames";
import usePanelParams from "context/usePanelParams";

interface Props {
  className?: string;
}

const EnrollClusterButton: FC<Props> = ({ className }) => {
  const panelParams = usePanelParams();

  return (
    <ActionButton
      onClick={panelParams.openEnrollCluster}
      appearance="positive"
      className={classnames("u-no-margin--bottom has-icon", className)}
    >
      <Icon name="plus" light />
      <span>Enroll cluster</span>
    </ActionButton>
  );
};

export default EnrollClusterButton;

import { ActionButton, Icon } from "@canonical/react-components";
import { FC } from "react";
import classnames from "classnames";
import usePanelParams from "context/usePanelParams";

interface Props {
  className?: string;
}

const EnrolClusterButton: FC<Props> = ({ className }) => {
  const panelParams = usePanelParams();

  return (
    <ActionButton
      onClick={panelParams.openEnrolCluster}
      appearance="positive"
      className={classnames("u-no-margin--bottom has-icon", className)}
    >
      <Icon name="plus" light />
      <span>Enrol cluster</span>
    </ActionButton>
  );
};

export default EnrolClusterButton;

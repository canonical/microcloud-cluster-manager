import { Button, Icon } from "@canonical/react-components";
import type { FC } from "react";
import usePanelParams from "context/usePanelParams";

const CreateClusterLinkButton: FC = () => {
  const panelParams = usePanelParams();

  return (
    <Button
      appearance="positive"
      className="u-no-margin--bottom has-icon"
      onClick={() => {
        panelParams.openCreateClusterLink();
      }}
      hasIcon
    >
      <Icon name="plus" light />
      <span>Create cluster link</span>
    </Button>
  );
};

export default CreateClusterLinkButton;

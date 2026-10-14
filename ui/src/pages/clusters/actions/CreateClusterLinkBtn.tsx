import { Button, Icon } from "@canonical/react-components";
import type { FC } from "react";
import usePanelParams from "context/usePanelParams";

const CreateClusterLinkBtn: FC = () => {
  const panelParams = usePanelParams();

  return (
    <div>
      <Button
        appearance=""
        className="u-no-margin--bottom has-icon"
        onClick={() => {
          panelParams.openCreateClusterLink();
        }}
        hasIcon
      >
        <Icon name="plus" />
        <span>Create Cluster Link</span>
      </Button>
    </div>
  );
};

export default CreateClusterLinkBtn;

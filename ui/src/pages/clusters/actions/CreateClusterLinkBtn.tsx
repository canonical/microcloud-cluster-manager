import { Button, Icon } from "@canonical/react-components";
import type { FC } from "react";
import usePanelParams from "context/usePanelParams";

interface Props {
  clusterName?: string;
}

const CreateClusterLinkBtn: FC<Props> = ({ clusterName }) => {
  const panelParams = usePanelParams();

  return (
    <Button
      appearance="positive"
      className="u-no-margin--bottom has-icon"
      onClick={() => {
        panelParams.openCreateClusterLink(clusterName);
      }}
      hasIcon
    >
      <Icon name="plus" className={"is-light"} />
      <span>Create cluster link</span>
    </Button>
  );
};

export default CreateClusterLinkBtn;

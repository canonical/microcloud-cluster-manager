import { Button, Icon } from "@canonical/react-components";
import type { FC } from "react";
import usePanelParams from "context/usePanelParams";

interface Props {
  cluster?: string;
}

const CreateClusterLinkBtn: FC<Props> = ({ cluster }) => {
  const panelParams = usePanelParams();

  return (
    <Button
      appearance="positive"
      className="u-no-margin--bottom has-icon"
      onClick={() => {
        panelParams.openCreateClusterLink(cluster);
      }}
      hasIcon
    >
      <Icon name="plus" className={"is-light"} />
      <span>Create cluster link</span>
    </Button>
  );
};

export default CreateClusterLinkBtn;

import { Button, Icon } from "@canonical/react-components";
import type { FC } from "react";
import usePanelParams from "context/usePanelParams";
import type { Cluster } from "types/cluster";

interface Props {
  cluster: Cluster;
}

const CreateClusterLinkBtn: FC<Props> = ({ cluster }) => {
  const panelParams = usePanelParams();

  return (
    <div>
      <Button
        appearance=""
        className="u-no-margin--bottom has-icon"
        onClick={() => {
          panelParams.openCreateClusterLink(cluster.name);
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

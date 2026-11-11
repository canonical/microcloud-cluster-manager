import { Button, Icon } from "@canonical/react-components";
import type { FC } from "react";
import usePanelParams from "context/usePanelParams";
import type { Cluster, ClusterLink } from "types/cluster";

interface Props {
  cluster: Cluster;
  link: ClusterLink;
}

const EditClusterLinkBtn: FC<Props> = ({ cluster, link }) => {
  const panelParams = usePanelParams();

  return (
    <Button
      appearance="base"
      className="u-no-margin--bottom has-icon"
      onClick={() => {
        panelParams.openEditClusterLink(cluster.name, link.name);
      }}
      hasIcon
    >
      <Icon name="edit" />
      <span className="u-off-screen">Edit {link.name}</span>
    </Button>
  );
};

export default EditClusterLinkBtn;

import { Button, Icon } from "@canonical/react-components";
import type { FC } from "react";
import usePanelParams from "context/usePanelParams";
import type { Cluster, ClusterLink } from "types/cluster";

interface Props {
  cluster: Cluster;
  link: ClusterLink;
  hasLabel?: boolean;
}

const EditClusterLinkBtn: FC<Props> = ({ cluster, link, hasLabel = true }) => {
  const panelParams = usePanelParams();

  return (
    <Button
      appearance={hasLabel ? "default" : "base"}
      className="u-no-margin--bottom"
      hasIcon
      title={hasLabel ? undefined : "Edit"}
      aria-label={hasLabel ? undefined : `Edit cluster link ${link.name}`}
      onClick={() => {
        panelParams.openEditClusterLink(cluster.name, link.name);
      }}
    >
      <Icon name="edit" />
      {hasLabel && <span>Edit</span>}
    </Button>
  );
};

export default EditClusterLinkBtn;

import { Button, Icon } from "@canonical/react-components";
import React, { FC } from "react";
import usePanelParams from "context/usePanelParams";
import { FIELD_DESCRIPTION } from "pages/clusters/ConfigureClusterPanel";

interface Props {
  clusterName: string;
  appearance?: "base" | "";
  label?: string;
}

const ClusterEditDescriptionBtn: FC<Props> = ({
  clusterName,
  appearance,
  label,
}) => {
  const panelParams = usePanelParams();

  return (
    <Button
      appearance={appearance}
      onClick={() =>
        panelParams.openConfigureCluster(clusterName, FIELD_DESCRIPTION)
      }
      hasIcon
    >
      <Icon name="edit" />
      {label && <span>{label}</span>}
    </Button>
  );
};

export default ClusterEditDescriptionBtn;

import { Button, Icon } from "@canonical/react-components";
import type { FC } from "react";
import usePanelParams from "context/usePanelParams";
import { pluralize } from "util/helpers";

interface Props {
  clusterNames: string[];
  onStart: () => void;
  onFinish: () => void;
}

const BulkConfigureClusterButton: FC<Props> = ({ clusterNames }) => {
  const panelParams = usePanelParams();

  return (
    <Button
      appearance=""
      className={"u-no-margin--bottom p-segmented-control__button"}
      hasIcon
      onClick={() => {
        panelParams.openBulkConfigureCluster(clusterNames);
      }}
    >
      <Icon name="settings" />
      <span>Configure {pluralize("cluster", clusterNames.length)}</span>
    </Button>
  );
};

export default BulkConfigureClusterButton;

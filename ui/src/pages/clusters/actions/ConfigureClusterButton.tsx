import { Button, Icon } from "@canonical/react-components";
import type { FC } from "react";
import usePanelParams from "context/usePanelParams";
import type { Cluster } from "types/cluster";
import classnames from "classnames";

interface Props {
  cluster: Cluster;
  className?: string;
  onClose?: () => void;
}

const ConfigureClusterButton: FC<Props> = ({ cluster, className, onClose }) => {
  const panelParams = usePanelParams();

  return (
    <Button
      className={classnames("u-no-margin--bottom", className)}
      hasIcon
      onClick={() => {
        panelParams.openConfigureCluster(cluster.name);
        onClose?.();
      }}
    >
      <Icon name="settings" />
      <span>Configure</span>
    </Button>
  );
};

export default ConfigureClusterButton;

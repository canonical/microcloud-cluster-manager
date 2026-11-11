import { Button, Icon } from "@canonical/react-components";
import type { FC } from "react";
import usePanelParams from "context/usePanelParams";
import type { Cluster } from "types/cluster";
import classnames from "classnames";

interface Props {
  cluster: Cluster;
  className?: string;
  onClose?: () => void;
  hasLabel?: boolean;
}

const ConfigureClusterButton: FC<Props> = ({
  cluster,
  className,
  onClose,
  hasLabel = true,
}) => {
  const panelParams = usePanelParams();

  return (
    <Button
      appearance={hasLabel ? "default" : "base"}
      className={classnames("u-no-margin--bottom", className)}
      hasIcon
      title={hasLabel ? undefined : "Edit"}
      aria-label={hasLabel ? undefined : `Edit cluster ${cluster.name}`}
      onClick={() => {
        panelParams.openConfigureCluster(cluster.name);
        onClose?.();
      }}
    >
      <Icon name={hasLabel ? "settings" : "edit"} />
      {hasLabel && <span>Configure</span>}
    </Button>
  );
};

export default ConfigureClusterButton;

import type { FC } from "react";
import { cloneElement } from "react";
import type { Cluster } from "types/cluster";
import ConfigureClusterButton from "pages/clusters/actions/ConfigureClusterButton";
import RemoveClusterButton from "pages/clusters/actions/RemoveClusterButton";
import ClusterUiButton from "pages/clusters/actions/ClusterUiButton";
import ClusterMetricsButton from "pages/clusters/actions/ClusterMetricsButton";
import { ContextualMenu } from "@canonical/react-components";

interface Props {
  cluster: Cluster;
}

const ClusterActions: FC<Props> = ({ cluster }) => {
  const menuElements = [
    <ConfigureClusterButton
      cluster={cluster}
      className="p-contextual-menu__link"
      key="configure"
    />,
    <ClusterUiButton
      uiUrl={cluster.ui_url}
      className="p-contextual-menu__link"
      key="ui"
    />,
    <ClusterMetricsButton
      clusterName={cluster.name}
      className="p-contextual-menu__link"
      key="metrics"
    />,
    <RemoveClusterButton
      clusterName={cluster.name}
      className="p-contextual-menu__link"
      key="remove"
    />,
  ];

  return (
    <ContextualMenu
      closeOnOutsideClick={false}
      toggleLabel=""
      position="left"
      hasToggleIcon
      title="actions"
      toggleAppearance="base"
    >
      {(close: () => void) => (
        <span>
          {[...menuElements].map((item) =>
            cloneElement(item, { onClose: close }),
          )}
        </span>
      )}
    </ContextualMenu>
  );
};

export default ClusterActions;

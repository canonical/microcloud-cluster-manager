import type { FC } from "react";
import { cloneElement } from "react";
import type { Cluster } from "types/cluster";
import ConfigureClusterBtn from "pages/clusters/actions/ConfigureClusterBtn";
import RemoveClusterBtn from "pages/clusters/actions/RemoveClusterBtn";
import ClusterUiBtn from "pages/clusters/actions/ClusterUiBtn";
import ClusterMetricsBtn from "pages/clusters/actions/ClusterMetricsBtn";
import { ContextualMenu } from "@canonical/react-components";

interface Props {
  cluster: Cluster;
}

const ClusterActions: FC<Props> = ({ cluster }) => {
  const menuElements = [
    <ConfigureClusterBtn
      cluster={cluster}
      className="p-contextual-menu__link"
      key="configure"
    />,
    <ClusterUiBtn
      lxdUrl={cluster.lxd_url}
      className="p-contextual-menu__link"
      key="ui"
    />,
    <ClusterMetricsBtn
      clusterName={cluster.name}
      className="p-contextual-menu__link"
      key="metrics"
    />,
    <RemoveClusterBtn
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
      className="cluster-actions"
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

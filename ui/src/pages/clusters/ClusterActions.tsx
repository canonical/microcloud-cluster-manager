import type { FC } from "react";
import { ContextualMenu, Icon } from "@canonical/react-components";
import type { Cluster } from "types/cluster";
import RemoveClusterButton from "pages/clusters/actions/RemoveClusterButton";
import ClusterMetricsButton from "pages/clusters/actions/ClusterMetricsButton";
import ClusterUiButton from "pages/clusters/actions/ClusterUiButton";
import ConfigureClusterButton from "pages/clusters/actions/ConfigureClusterButton";

interface Props {
  cluster: Cluster;
}

const ClusterActions: FC<Props> = ({ cluster }) => {
  return (
    <ContextualMenu
      position="right"
      scrollOverflow
      hasToggleIcon={false}
      toggleAppearance="base"
      toggleClassName="u-no-margin--bottom has-icon"
      toggleLabel={<Icon name="chevron-down" />}
      toggleProps={{ "aria-label": `Actions for cluster ${cluster.name}` }}
      dropdownClassName="cluster-actions-menu"
    >
      {(close) => (
        <>
          <ConfigureClusterButton
            cluster={cluster}
            className="p-contextual-menu__link"
            onClose={close}
          />
          <ClusterUiButton
            lxdUrl={cluster.lxd_url}
            className="p-contextual-menu__link"
            onClose={close}
          />
          <ClusterMetricsButton
            clusterName={cluster.name}
            className="p-contextual-menu__link"
            onClose={close}
          />
          <RemoveClusterButton
            clusterName={cluster.name}
            className="p-contextual-menu__link"
            onClose={close}
          />
        </>
      )}
    </ContextualMenu>
  );
};

export default ClusterActions;

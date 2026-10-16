import type { FC } from "react";
import { Button, Icon } from "@canonical/react-components";
import type { Cluster } from "types/cluster";
import usePanelParams from "context/usePanelParams";
import ActionsList from "components/ActionsList";
import RemoveClusterButton from "pages/clusters/actions/RemoveClusterButton";

interface Props {
  cluster: Cluster;
}

const ClusterActions: FC<Props> = ({ cluster }) => {
  const panelParams = usePanelParams();

  return (
    <ActionsList
      items={[
        <Button
          appearance="base"
          hasIcon
          className="u-no-margin--bottom"
          title="Edit"
          aria-label={`Edit cluster ${cluster.name}`}
          key="edit"
          onClick={() => {
            panelParams.openConfigureCluster(cluster.name);
          }}
        >
          <Icon name="edit" />
        </Button>,
        <RemoveClusterButton
          clusterName={cluster.name}
          hasLabel={false}
          key="remove"
        />,
      ]}
    />
  );
};

export default ClusterActions;

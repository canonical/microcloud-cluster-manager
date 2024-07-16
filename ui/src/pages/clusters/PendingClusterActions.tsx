import { FC } from "react";
import DeleteClusterButton from "./DeleteClusterButton";
import { Cluster } from "types/cluster";
import ApproveClusterButton from "./ApproveClusterButton";

interface Props {
  cluster: Cluster;
}

const PendingClusterActions: FC<Props> = ({ cluster }) => {
  return (
    <div className="cluster-actions">
      <ApproveClusterButton clusterName={cluster.name} />
      <DeleteClusterButton clusterName={cluster.name} />
    </div>
  );
};

export default PendingClusterActions;

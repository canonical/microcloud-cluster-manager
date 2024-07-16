import { Icon } from "@canonical/react-components";
import { FC } from "react";
import { Cluster } from "types/cluster";
import { getMinutesSinceLastHeartbeat } from "util/helpers";

interface Props {
  cluster: Cluster;
}

const ClusterStatus: FC<Props> = ({ cluster }: Props) => {
  const lastHeartbeatMins = getMinutesSinceLastHeartbeat(cluster);
  const heartbeatClass =
    lastHeartbeatMins < 5 ? "succeeded-small" : "failed-small";

  return (
    <div>
      <Icon
        className="is-light p-side-navigation__icon"
        name={`status-${heartbeatClass}`}
      />
      {/* Status displays as Online when the cluster was last seen less than 5 minutes ago. */}
      {lastHeartbeatMins < 5 ? "Online" : "Offline"}
    </div>
  );
};

export default ClusterStatus;

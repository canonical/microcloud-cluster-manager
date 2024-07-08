import { Icon } from "@canonical/react-components";
import { FC } from "react";
import { Cluster } from "types/cluster";
import { getHeartbeatStatus } from "util/helpers";

interface Props {
  cluster: Cluster;
}

const ClusterHeartbeat: FC<Props> = ({ cluster }: Props) => {
  const heartbeatStatus = getHeartbeatStatus(cluster);
  const heartbeatClass =
    heartbeatStatus != "Unresponsive" ? "succeeded-small" : "failed-small";

  return (
    <div>
      <Icon
        className="is-light p-side-navigation__icon"
        name={`status-${heartbeatClass}`}
      />
      {heartbeatStatus}
    </div>
  );
};

export default ClusterHeartbeat;

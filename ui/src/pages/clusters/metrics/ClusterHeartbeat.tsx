import type { FC } from "react";
import type { Cluster } from "types/cluster";
import { getMinutesSinceLastHeartbeat, isoTimeToString } from "util/helpers";
import { getLastSeenLabel } from "util/lastSeen";

interface Props {
  cluster: Cluster;
}

const ClusterHeartbeat: FC<Props> = ({ cluster }: Props) => {
  const lastHeartbeatMins = getMinutesSinceLastHeartbeat(cluster);

  return (
    <div
      title={isoTimeToString(cluster.last_status_update_at)}
      className="u-text--muted"
    >
      {getLastSeenLabel(lastHeartbeatMins)}
    </div>
  );
};

export default ClusterHeartbeat;

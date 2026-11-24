import StatusIcon from "components/StatusIcon";
import type { FC } from "react";
import type { Cluster } from "types/cluster";
import { getMinutesSinceLastHeartbeat } from "util/helpers";

interface Props {
  cluster: Cluster;
}

const ClusterStatus: FC<Props> = ({ cluster }: Props) => {
  const lastHeartbeatMins = getMinutesSinceLastHeartbeat(cluster);
  const tone = lastHeartbeatMins < 5 ? "positive" : "negative";

  return (
    <div>
      <StatusIcon tone={tone} />
      {/* Status displays as Online when the cluster was last seen less than 5 minutes ago. */}
      {lastHeartbeatMins < 5 ? "Online" : "Offline"}
    </div>
  );
};

export default ClusterStatus;

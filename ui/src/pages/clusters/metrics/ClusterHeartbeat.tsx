import { FC } from "react";
import { Cluster } from "types/cluster";
import { getMinutesSinceLastHeartbeat, isoTimeToString } from "util/helpers";

interface Props {
  cluster: Cluster;
}

const ClusterHeartbeat: FC<Props> = ({ cluster }: Props) => {
  const lastHeartbeatMins = getMinutesSinceLastHeartbeat(cluster);
  const lastHeartbeatHrs = Math.floor(lastHeartbeatMins / 3600000);
  let returnStr;

  if (lastHeartbeatMins <= 1) {
    returnStr = `seen 1 minute ago`;
  } else if (lastHeartbeatMins < 5) {
    returnStr = `seen ${lastHeartbeatMins} minutes ago`;
  } else {
    returnStr =
      lastHeartbeatHrs < 1 //Displayed for "Last Seen"'s of 5-59 Minutes
        ? `seen in last hour`
        : lastHeartbeatHrs < 2
          ? `seen 1 hour ago`
          : `seen ${lastHeartbeatHrs} hours ago`;
  }

  return (
    <div
      title={isoTimeToString(cluster.last_status_update_at)}
      className="u-text--muted"
    >
      {returnStr}
    </div>
  );
};

export default ClusterHeartbeat;

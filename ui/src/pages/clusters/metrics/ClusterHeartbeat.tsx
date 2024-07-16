import { FC } from "react";
import { Cluster } from "types/cluster";
import { getMinutesSinceLastHeartbeat } from "util/helpers";

interface Props {
  cluster: Cluster;
}

const ClusterHeartbeat: FC<Props> = ({ cluster }: Props) => {
  const lastHeartbeatMins = getMinutesSinceLastHeartbeat(cluster);
  const lastHeartbeatHrs = Math.floor(lastHeartbeatMins / 3600000);
  let returnStr = "";

  if (lastHeartbeatMins <= 1) {
    returnStr = `< 1 minute ago`;
  } else if (lastHeartbeatMins < 5) {
    returnStr = `${lastHeartbeatMins} minutes ago`;
  } else {
    returnStr =
      lastHeartbeatHrs < 1 //Displayed for "Last Seen"'s of 5-59 Minutes
        ? `< 1 hour ago`
        : lastHeartbeatHrs < 2
          ? `1 hour ago`
          : `${lastHeartbeatHrs} hours ago`;
  }

  return <div>{returnStr}</div>;
};

export default ClusterHeartbeat;

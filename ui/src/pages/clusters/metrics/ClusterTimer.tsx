import { FC, useEffect, useState } from "react";
import { Cluster } from "types/cluster";
import { getSecondsSinceLastHeartbeat } from "util/helpers";

interface Props {
  cluster: Cluster;
}

const ClusterTimer: FC<Props> = ({ cluster }: Props) => {
  const [seconds, setSeconds] = useState(getSecondsSinceLastHeartbeat(cluster));

  useEffect(() => {
    const timerId = setInterval(() => {
      setSeconds(getSecondsSinceLastHeartbeat(cluster));
    }, 1000);
    return () => clearInterval(timerId);
  }, [cluster, seconds]);

  const getFormattedTimeLeft = () => {
    const minutes = Math.floor(seconds / 60);
    const displaySeconds = seconds % 60;

    return `${minutes}:${displaySeconds < 10 ? `0${displaySeconds}` : displaySeconds}`;
  };

  return (
    <div className="cluster-detail-countdown">
      <div className="u-no-margin u-no-padding p-heading--3">
        {getFormattedTimeLeft()}
      </div>
      <div>Since last heartbeat</div>
    </div>
  );
};

export default ClusterTimer;

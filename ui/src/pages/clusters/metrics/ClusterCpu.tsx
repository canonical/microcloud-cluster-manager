import { FC } from "react";
import { Cluster } from "types/cluster";
import classnames from "classnames";

interface Props {
  cluster: Cluster;
  containerClassname?: string;
}

export const ClusterCpu: FC<Props> = ({
  cluster,
  containerClassname,
}: Props) => {
  const averageReadings = [
    cluster.cpu_load_1,
    cluster.cpu_load_5,
    cluster.cpu_load_15,
  ];

  return (
    <div className={classnames("cpu-badges", containerClassname)}>
      {averageReadings.map((reading, index) => (
        <div key={index} className={"cpu-badge"}>
          {reading}
        </div>
      ))}
    </div>
  );
};

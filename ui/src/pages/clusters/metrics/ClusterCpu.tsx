import { FC } from "react";
import { Cluster } from "types/cluster";

interface Props {
  cluster: Cluster;
}

export const ClusterCpu: FC<Props> = ({ cluster }: Props) => {
  const averageReadings = [
    cluster.cpu_load_1,
    cluster.cpu_load_5,
    cluster.cpu_load_15,
  ];

  const getStyle = (reading: number) => {
    if (reading <= 0.6) {
      return "129, 70, 186";
    } else {
      return "199, 22, 43";
    }
  };

  return (
    <div className="cpu-badges">
      {averageReadings.map((reading, index) => (
        <div
          key={index}
          className={"cpu-badge"}
          style={{
            backgroundColor: `rgba(${getStyle(parseFloat(reading))}, ${parseFloat(reading)})`,
          }}
        >
          {reading}
        </div>
      ))}
    </div>
  );
};

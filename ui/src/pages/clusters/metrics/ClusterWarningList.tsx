import { FC } from "react";
import { Cluster } from "types/cluster";
import { getClusterWarnings } from "util/clusterWarnings";

interface Props {
  cluster: Cluster;
}

export const ClusterWarningList: FC<Props> = ({ cluster }: Props) => {
  const warnings = getClusterWarnings(cluster);

  return (
    <>
      <h5>Warnings</h5>
      {warnings.length > 0 ? (
        <ul className="cluster-warning-list">
          {warnings.map((warning, index) => (
            <li key={index}>{warning}</li>
          ))}
        </ul>
      ) : (
        <>
          <div>This cluster has no warnings.</div>
          <div className="u-text--muted">You’re doing something right!</div>
        </>
      )}
    </>
  );
};

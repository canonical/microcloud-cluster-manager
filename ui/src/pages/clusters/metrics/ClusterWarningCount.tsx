import type { FC } from "react";
import type { Cluster } from "types/cluster";
import { Icon, Tooltip } from "@canonical/react-components";
import { getClusterWarnings } from "util/clusterWarnings";

interface Props {
  cluster: Cluster;
}

export const ClusterWarningCount: FC<Props> = ({ cluster }: Props) => {
  const warnings = getClusterWarnings(cluster);
  const warningCount = warnings.length;

  return warningCount == 0 ? (
    0
  ) : (
    <Tooltip
      message={warnings.map((warning, index) => (
        <div key={index}>{warning}</div>
      ))}
      position="btm-center"
    >
      <div className="tooltip-toggle">
        {warningCount}
        {warningCount > 0 && (
          <>
            {" "}
            <Icon name="warning" />
          </>
        )}
      </div>
    </Tooltip>
  );
};

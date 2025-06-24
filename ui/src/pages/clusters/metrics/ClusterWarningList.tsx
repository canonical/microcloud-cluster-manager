import React, { FC } from "react";
import { Cluster } from "types/cluster";
import { getClusterWarnings } from "util/clusterWarnings";
import { EmptyState, Icon, Notification } from "@canonical/react-components";
import { pluralize } from "util/helpers";
import usePanelParams from "context/usePanelParams";
import { FIELD_MEMORY_THRESHOLD } from "pages/clusters/ConfigureClusterPanel";

interface Props {
  cluster: Cluster;
}

export const ClusterWarningList: FC<Props> = ({ cluster }: Props) => {
  const panelParams = usePanelParams();
  const warnings = getClusterWarnings(cluster);
  const isEmpty = warnings.length === 0;

  return isEmpty ? (
    <EmptyState
      className="empty-state"
      image={<Icon name="success-grey" className="empty-state-icon" />}
      title="No warnings"
    >
      <p>You’re doing something right!</p>
    </EmptyState>
  ) : (
    <div className="warning-list">
      <h2 className="p-heading--4">
        {warnings.length === 0 ? (
          <>No warnings</>
        ) : (
          <>
            {warnings.length} {pluralize("warning", warnings.length)}
          </>
        )}
      </h2>
      {warnings.length > 0 ? (
        warnings.map((warning, index) => {
          const isMemoryUsage = warning.startsWith("Memory usage");
          const isDiskUsage = warning.startsWith("Disk usage");
          const canConfigure = isMemoryUsage || isDiskUsage;

          const actions = [];
          if (canConfigure) {
            const focusField = isMemoryUsage
              ? FIELD_MEMORY_THRESHOLD
              : undefined;
            actions.push({
              label: "Configure threshold",
              onClick: () =>
                panelParams.openConfigureCluster(cluster.name, focusField),
            });
          }

          return (
            <Notification
              severity="caution"
              key={index}
              title={warning}
              actions={actions}
            />
          );
        })
      ) : (
        <>
          <div className="u-text--muted">You’re doing something right!</div>
        </>
      )}
    </div>
  );
};

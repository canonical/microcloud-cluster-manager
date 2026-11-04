import type { FC } from "react";
import { Icon } from "@canonical/react-components";
import type { Cluster } from "types/cluster";

interface Props {
  cluster: Cluster;
}

const STATUS_ICONS: Record<string, string> = {
  ACTIVE: "status-succeeded-small",
  ERROR: "status-failed-small",
  OFFLINE: "status-failed-small",
  PENDING: "status-queued-small",
};

const STATUS_LABELS: Record<string, string> = {
  ACTIVE: "Active",
  ERROR: "Error",
  OFFLINE: "Offline",
  PENDING: "Pending",
};

const ClusterStatus: FC<Props> = ({ cluster }) => {
  const status = cluster.status?.toUpperCase() || "PENDING";
  const icon = STATUS_ICONS[status] || "status-queued-small";
  const label = STATUS_LABELS[status] || status;

  return (
    <>
      <Icon name={icon} className="status-icon" />
      {label}
    </>
  );
};

export default ClusterStatus;

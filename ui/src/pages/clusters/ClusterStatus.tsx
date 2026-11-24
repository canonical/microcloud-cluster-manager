import type { FC } from "react";
import StatusIcon from "components/StatusIcon";
import type { Cluster } from "types/cluster";
import type { StatusTone } from "util/statusColors";

interface Props {
  cluster: Cluster;
}

const STATUS_ICONS: Record<string, StatusTone> = {
  ACTIVE: "positive",
  ERROR: "negative",
  OFFLINE: "negative",
  PENDING: "neutral",
};

const STATUS_LABELS: Record<string, string> = {
  ACTIVE: "Active",
  ERROR: "Error",
  OFFLINE: "Offline",
  PENDING: "Pending",
};

const ClusterStatus: FC<Props> = ({ cluster }) => {
  const status = cluster.status?.toUpperCase() || "PENDING";
  const icon = STATUS_ICONS[status] || "neutral";
  const label = STATUS_LABELS[status] || status;

  return (
    <>
      <StatusIcon tone={icon} className="status-icon" />
      {label}
    </>
  );
};

export default ClusterStatus;

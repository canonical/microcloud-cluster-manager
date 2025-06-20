import { Cluster } from "types/cluster";
import { getMinutesSinceLastHeartbeat, pluralize } from "util/helpers";

const DISK_USAGE_PERCENTAGE_THRESHOLD = 90;
const MEMORY_USAGE_PERCENTAGE_THRESHOLD = 90;

export const getClusterWarnings = (cluster: Cluster): string[] => {
  const result: string[] = [];

  const diskUsagePercent =
    (100 / cluster.disk_total_size) * cluster.disk_usage || 0;
  if (diskUsagePercent > DISK_USAGE_PERCENTAGE_THRESHOLD) {
    result.push(`Disk usage is at ${Math.ceil(diskUsagePercent)}%`);
  }

  const memoryUsagePercent =
    (100 / cluster.memory_total_amount) * cluster.memory_usage || 0;
  if (memoryUsagePercent > MEMORY_USAGE_PERCENTAGE_THRESHOLD) {
    result.push(`Memory usage is at ${Math.ceil(memoryUsagePercent)}%`);
  }

  const nonOnlineMemberCount =
    cluster.member_statuses
      .filter((item) => item.status !== "Online")
      .map((item) => item.count)
      .reduce((a, b) => a + b, 0) || 0;
  if (nonOnlineMemberCount > 0) {
    result.push(
      `${nonOnlineMemberCount} cluster ${pluralize("member", nonOnlineMemberCount)} not online`,
    );
  }

  const errorInstanceCount =
    cluster.instance_statuses.find((item) => item.status === "Error")?.count ||
    0;
  if (errorInstanceCount > 0) {
    result.push(
      `${errorInstanceCount} ${pluralize("instance", errorInstanceCount)} in error state`,
    );
  }

  const lastHeartbeatMins = getMinutesSinceLastHeartbeat(cluster);
  if (lastHeartbeatMins >= 5) {
    result.push(
      `Cluster has not sent a heartbeat in the last ${lastHeartbeatMins} minutes`,
    );
  }

  return result;
};

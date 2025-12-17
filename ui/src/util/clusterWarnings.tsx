import type { Cluster } from "types/cluster";
import { getMinutesSinceLastHeartbeat, pluralize } from "util/helpers";

export const getClusterWarnings = (cluster: Cluster): string[] => {
  const result: string[] = [];

  cluster.storage_pool_usages.forEach((pool) => {
    const usagePercent = (100 / pool.total) * pool.usage || 0;
    const memberSuffix = pool.member !== "" ? ` on ${pool.member}` : "";

    if (usagePercent > cluster.disk_threshold) {
      result.push(
        `Storage pool "${pool.name}${memberSuffix}" usage is at ${Math.ceil(usagePercent)}%`,
      );
    }
  });

  const memoryUsagePercent =
    (100 / cluster.memory_total_amount) * cluster.memory_usage || 0;
  if (memoryUsagePercent > cluster.memory_threshold) {
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

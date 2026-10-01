// The cluster connector stores the metrics of a cluster under a Prometheus job named after the cluster, with "-"
// replaced by "_" (see parsePrometheusMetrics). The Grafana dashboard filters on that job.
export const getClusterMetricsUrl = (
  grafanaBaseUrl: string,
  clusterName: string,
): string => {
  const job = clusterName.replaceAll("-", "_");
  return `${grafanaBaseUrl}/lxd?orgId=1&var-job=${encodeURIComponent(job)}`;
};

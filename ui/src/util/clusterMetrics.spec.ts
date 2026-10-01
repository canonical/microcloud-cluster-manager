import { getClusterMetricsUrl } from "./clusterMetrics";

describe("getClusterMetricsUrl", () => {
  it("uses the job name the cluster connector stores metrics under", () => {
    expect(getClusterMetricsUrl("https://grafana.example", "prod-east-1")).toBe(
      "https://grafana.example/lxd?orgId=1&var-job=prod_east_1",
    );
  });

  it("encodes the job name", () => {
    expect(getClusterMetricsUrl("https://grafana.example", "rack&2")).toBe(
      "https://grafana.example/lxd?orgId=1&var-job=rack%262",
    );
  });
});

import { expect, test } from "./fixtures/test-extension";
import { visitClusterLinks, waitForTunnel } from "./helpers/cluster-links";
import { tunnelCluster } from "./helpers/remote-clusters";

test("list the cluster links of a real cluster through the tunnel", async ({
  page,
}) => {
  await waitForTunnel(page, tunnelCluster);
  await visitClusterLinks(page, tunnelCluster);
  await expect(page.getByText("No cluster links found")).toBeVisible();
});

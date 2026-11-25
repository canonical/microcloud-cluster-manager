import { expect, test } from "./fixtures/test-extension";
import {
  createClusterLink,
  deleteAllClusterLinks,
  deleteClusterLink,
  editClusterLink,
  visitClusterLinks,
  waitForTunnel,
} from "./helpers/cluster-links";
import { tunnelCluster01, tunnelCluster02 } from "./helpers/remote-clusters";

test("list the cluster links of a real cluster through the tunnel", async ({
  page,
}) => {
  await Promise.all([
    waitForTunnel(page, tunnelCluster01),
    waitForTunnel(page, tunnelCluster02),
  ]);
  await deleteAllClusterLinks(page, [tunnelCluster01, tunnelCluster02]);
  await visitClusterLinks(page, tunnelCluster01);
  await expect(page.getByText("No cluster links found")).toBeVisible();

  const linkName = `${tunnelCluster02}-link`;
  await createClusterLink(page, tunnelCluster01, tunnelCluster02);
  await editClusterLink(page, linkName, tunnelCluster01);
  await deleteClusterLink(page, linkName, tunnelCluster01, tunnelCluster02);
});

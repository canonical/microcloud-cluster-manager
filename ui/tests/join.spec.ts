import { expect, test } from "./fixtures/test-extension";
import { visitClusterLinks } from "./helpers/cluster-links";
import { createRemoteClusterToken } from "./helpers/remote-cluster-tokens";
import { joinCluster } from "./helpers/remote-clusters";
import { runCommand } from "./helpers/shell";

// Joining uses up the token, so a retry cannot succeed.
test.describe.configure({ retries: 0 });

test("join a MicroCloud with the displayed command", async ({ page }) => {
  await createRemoteClusterToken(page, joinCluster);
  const command = (await page.locator(".command").innerText()).trim();

  const output = runCommand(
    `lxc --project microcloud-mcm exec --env TEST_CONSOLE=0 ${joinCluster} -- ${command}`,
  );
  expect(output).toContain("Successfully joined cluster manager");

  // A joined cluster only registers a tunnel after its reverse tunnel is enabled.
  await visitClusterLinks(page, joinCluster);
  await expect(
    page.getByText("Cluster links require a registered tunnel."),
  ).toBeVisible();
});

import { test } from "@playwright/test";
import { ensureClusterExists } from "./helpers/remote-clusters";

test("ensure cluster is present", async ({ page }) => {
  const clusterName = "cluster-01";
  await ensureClusterExists(page, clusterName);
});

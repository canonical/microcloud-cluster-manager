import { test } from "@playwright/test";
import {
  ensureClusterExists,
  setClusterDiskLimit,
} from "./helpers/remote-clusters";

test("find cluster and set disk limit", async ({ page }) => {
  const cluster = "cluster-01";
  await ensureClusterExists(page, cluster);
  await setClusterDiskLimit(page, cluster, "42");
});

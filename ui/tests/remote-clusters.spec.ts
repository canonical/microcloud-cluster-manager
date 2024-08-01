import { test, expect } from "@playwright/test";
import {
  approvePendingCluster,
  checkClusterExistInTable,
  deletePendingCluster,
} from "./helpers/remote-clusters";

test("approve a pending cluster", async ({ page }) => {
  const clusterName = await approvePendingCluster(page);
  const clusterIsActive = await checkClusterExistInTable(
    page,
    clusterName,
    "Active",
  );
  expect(clusterIsActive).toBe(true);
});

test("delete pending cluster", async ({ page }) => {
  const clusterName = await deletePendingCluster(page);
  const clusterExists = await checkClusterExistInTable(
    page,
    clusterName,
    "Pending",
  );
  expect(clusterExists).toBe(false);
});

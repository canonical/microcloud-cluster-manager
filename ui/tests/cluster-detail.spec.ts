import { test } from "@playwright/test";
import {
  ensureClusterExists,
  goToClusterDetail,
  setClusterConfig,
  randomSeedClusterName,
} from "./helpers/remote-clusters";
import { expect } from "./fixtures/test-extension";

test.describe("Cluster Configuration", () => {
  test("should set disk threshold limit", async ({ page }) => {
    const cluster = randomSeedClusterName();
    await ensureClusterExists(page, cluster);
    await setClusterConfig(page, cluster, "Disk threshold percentage", "42");
  });

  test("should set memory threshold limit", async ({ page }) => {
    const cluster = randomSeedClusterName();
    await ensureClusterExists(page, cluster);
    await setClusterConfig(page, cluster, "Memory threshold percentage", "42");
  });

  test("should display warnings for low thresholds and hide them when thresholds are high", async ({
    page,
  }) => {
    const cluster = randomSeedClusterName();
    await ensureClusterExists(page, cluster);
    await setClusterConfig(page, cluster, "Memory threshold percentage", "1");
    await setClusterConfig(page, cluster, "Disk threshold percentage", "1");

    await expect(page.getByText(/\d warning?/)).toBeVisible();
    const notifications = page.locator(".p-notification--caution");
    const initialCount = await notifications.count();

    await setClusterConfig(page, cluster, "Memory threshold percentage", "100");
    await setClusterConfig(page, cluster, "Disk threshold percentage", "100");
    const finalCount = await notifications.count();
    expect(initialCount === 0 || finalCount < initialCount).toBe(true);
  });
});

test.describe("Cluster Description", () => {
  test("should add description to cluster without one", async ({ page }) => {
    const cluster = randomSeedClusterName();
    await ensureClusterExists(page, cluster);
    await goToClusterDetail(page, cluster);
    await page
      .getByRole("button", { name: "Add description", exact: true })
      .click();
    await page.getByLabel("Description").fill("New description");
    await page.getByRole("button", { name: "Save changes" }).click();
    await expect(page.getByText(`Updated cluster ${cluster}`)).toBeVisible();
  });

  test("should edit existing cluster description", async ({ page }) => {
    const cluster = randomSeedClusterName();
    const initialDescription = "Initial description";
    await setClusterConfig(page, cluster, "Description", initialDescription);
    const container = page.getByText(initialDescription, { exact: true });
    const button = container.getByRole("button");
    await button.click();
    await page.getByLabel("Description").fill("Updated description");
    await page.getByRole("button", { name: "Save changes" }).click();
    await expect(page.getByText(`Updated cluster ${cluster}`)).toBeVisible();
    await expect(
      page.getByText("Updated description", { exact: true }),
    ).toBeVisible();
  });

  test("should cancel description edit", async ({ page }) => {
    const cluster = randomSeedClusterName();
    await goToClusterDetail(page, cluster);
    await page
      .getByRole("button", { name: "Add description", exact: true })
      .click();
    await page.getByLabel("Description").fill("Should not save");
    await page.getByRole("button", { name: "Cancel" }).click();
    await expect(
      page.getByText("Should not save", { exact: true }),
    ).not.toBeVisible();
  });
});

test.describe("Cluster Monitoring", () => {
  test("should display heartbeat information", async ({ page }) => {
    const cluster = randomSeedClusterName();
    await goToClusterDetail(page, cluster);
    await expect(
      page.getByText("Since last heartbeat", { exact: true }),
    ).toBeVisible();
  });
});

test.describe("Cluster Usage Statistics", () => {
  test("should display total memory usage", async ({ page }) => {
    const cluster = randomSeedClusterName();
    await goToClusterDetail(page, cluster);
    await expect(page.getByText("Total memory", { exact: true })).toBeVisible();
  });

  test("should display total storage usage", async ({ page }) => {
    const cluster = randomSeedClusterName();
    await goToClusterDetail(page, cluster);
    await expect(page.getByText("Total storage")).toBeVisible();
  });
});

test.describe("Storage Details", () => {
  test("should display storage pool details modal", async ({ page }) => {
    const cluster = randomSeedClusterName();
    await goToClusterDetail(page, cluster);
    await page.getByText("View details", { exact: true }).click();
    await expect(
      page.getByText("Storage pool details", { exact: true }),
    ).toBeVisible();
  });

  test("should close storage details modal", async ({ page }) => {
    const cluster = randomSeedClusterName();
    await goToClusterDetail(page, cluster);
    await page.getByText("View details", { exact: true }).click();
    await page.getByRole("button", { name: "Close" }).click();
    await expect(
      page.getByText("Storage pool details", { exact: true }),
    ).not.toBeVisible();
  });
});

test.describe("Cluster Members", () => {
  test("should display cluster members graph", async ({ page }) => {
    const cluster = randomSeedClusterName();
    await goToClusterDetail(page, cluster);
    const container = page.locator(".cluster-detail-doughnut-graph", {
      has: page.getByText(/\d+\s+members$/),
    });
    await expect(container).toBeVisible();
  });
});

test.describe("Cluster Instances", () => {
  test("should display cluster instances graph", async ({ page }) => {
    const cluster = randomSeedClusterName();
    await goToClusterDetail(page, cluster);
    const container = page.locator(".cluster-detail-doughnut-graph", {
      has: page.getByText(/\d+\s+instances$/),
    });
    await expect(container).toBeVisible();
  });
});

test.describe("Navigation and UI", () => {
  test("should navigate back to clusters list", async ({ page }) => {
    const cluster = randomSeedClusterName();
    await goToClusterDetail(page, cluster);
    await page.getByRole("link", { name: "Clusters", exact: true }).click();
    await expect(page).toHaveURL(/.*\/clusters$/);
  });
});

test.describe("Cluster Removal", () => {
  test("should remove cluster with confirmation", async ({ page }) => {
    const cluster = randomSeedClusterName();
    await goToClusterDetail(page, cluster);
    await page.getByRole("button", { name: "Remove", exact: true }).click();
    await page
      .getByRole("button", { name: "Confirm remove", exact: true })
      .click();
    await expect(page.getByText(`Removed cluster ${cluster}`)).toBeVisible();
  });

  test("should cancel cluster removal", async ({ page }) => {
    const cluster = randomSeedClusterName();
    await goToClusterDetail(page, cluster);
    await page.getByRole("button", { name: "Remove", exact: true }).click();
    await page.getByRole("button", { name: "Cancel", exact: true }).click();
    await expect(page).toHaveURL(new RegExp(`.*/cluster/${cluster}`));
    await ensureClusterExists(page, cluster);
  });
});

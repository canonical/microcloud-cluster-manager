import { expect, test } from "@playwright/test";

test("cluster list", async ({ page }) => {
  await page.goto("/ui/clusters");
  expect(await page.title()).toBe("LXD Cluster Manager");
});

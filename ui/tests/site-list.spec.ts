import { expect, test } from "@playwright/test";

test("site list", async ({ page }) => {
  await page.goto("/ui");
  expect(await page.title()).toBe("LXD Site Manager");
});

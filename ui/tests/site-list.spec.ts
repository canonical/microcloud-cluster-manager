import { expect, test } from "@playwright/test";

test("site list", async ({ page }) => {
  await page.goto("/");
  expect(await page.title()).toBe("LXD Site Manager");
});

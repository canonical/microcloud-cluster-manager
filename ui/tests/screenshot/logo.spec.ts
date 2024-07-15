import { test, expect } from "@playwright/test";

test("logo screenshot", async ({ page }) => {
  await page.goto("/ui/sites");
  const logo = page.locator(".l-navigation__drawer .p-panel__logo-image");
  await expect(logo).toHaveScreenshot();
});

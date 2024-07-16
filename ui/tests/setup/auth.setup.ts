import { Page, expect, test as setup } from "@playwright/test";
import { authFile } from "../fixtures/constants";

const loginUser = async (page: Page, sso: boolean) => {
  await page.getByRole("link", { name: "Login" }).click();
  if (sso) {
    await page.getByLabel("Email address*").click();
    await page.getByLabel("Email address*").fill(process.env.OIDC_USER || "");
    await page.getByLabel("Password*").click();
    await page.getByLabel("Password*").fill(process.env.OIDC_PASSWORD || "");
    await page.getByRole("button", { name: "Continue", exact: true }).click();
  }
  await expect(page.getByText("Log out")).toBeVisible();
};

setup("authenticate", async ({ page }) => {
  await page.goto("/ui");
  await loginUser(page, true);
  // Check logout functionality
  await page.getByText("Log out").click();
  await loginUser(page, false);

  await page.context().storageState({ path: authFile });
});

import { test, expect } from "@playwright/test";
import { goToSettingsPage, validateSettingValue } from "./helpers/settings";

const SETTINGS = [
  "API Version",
  "Cluster Connector Domain",
  "Cluster Connector Port",
  "OIDC Client ID",
  "OIDC Issuer",
  "OIDC Audience",
];

test("all settings exist", async ({ page }) => {
  await goToSettingsPage(page);
  for (const setting of SETTINGS) {
    const settingRow = page.getByText(setting);
    await expect(settingRow).toBeVisible();
  }
});

test("validate specific setting values", async ({ page }) => {
  await goToSettingsPage(page);
  await validateSettingValue(page, "API Version", "1.0");
});

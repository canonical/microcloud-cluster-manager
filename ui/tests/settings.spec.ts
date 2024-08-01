import { test, expect } from "@playwright/test";
import {
  goToSettingsPage,
  updateManagerSetting,
  updateMemberSetting,
  validateSettingValue,
} from "./helpers/settings";

const MANAGER_SETTINGS = [
  "oidc.issuer",
  "oidc.client.id",
  "oidc.audience",
  "global.address",
];

const MEMBER_SETTINGS = ["https_address", "external_address"];

// NOTE: this is an assumption that the manager system in the testing environment has a member with the name "member1"
const MEMBER_NAME = "member1";

test("all manager settings exists", async ({ page }) => {
  await goToSettingsPage(page);
  for (const setting of MANAGER_SETTINGS) {
    const settingRow = page.getByText(setting);
    await expect(settingRow).toBeVisible();
  }
});

test("all member settings exists", async ({ page }) => {
  await goToSettingsPage(page);
  const memberConfigRows = page
    .getByRole("row")
    .filter({ hasText: MEMBER_NAME });
  await expect(memberConfigRows).toHaveCount(2);
  await expect(memberConfigRows.nth(0)).toContainText(
    `${MEMBER_NAME}.${MEMBER_SETTINGS[0]}`,
  );
  await expect(memberConfigRows.nth(1)).toContainText(
    `${MEMBER_NAME}.${MEMBER_SETTINGS[1]}`,
  );
});

// NOTE: we DO NOT want to edit the oidc settings since it will break the login flow
test("edit manager setting global.address", async ({ page }) => {
  await goToSettingsPage(page);
  const settingValue = "http://localhost:8080";
  const settingRow = await updateManagerSetting(
    page,
    "global.address",
    settingValue,
  );
  await validateSettingValue(settingRow, settingValue);
});

test("edit member settings", async ({ page }) => {
  await goToSettingsPage(page);
  let settingValue = "0.0.0.0:9801";
  let settingRow = await updateMemberSetting(
    page,
    MEMBER_NAME,
    "https_address",
    settingValue,
  );
  await validateSettingValue(settingRow, settingValue);

  settingValue = "0.0.0.0:9810";
  settingRow = await updateMemberSetting(
    page,
    MEMBER_NAME,
    "external_address",
    settingValue,
  );
  await validateSettingValue(settingRow, settingValue);
});

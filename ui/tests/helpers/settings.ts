import type { Page } from "@playwright/test";
import { expect } from "../fixtures/test-extension";
import { gotoURL } from "./navigate";

export const goToSettingsPage = async (page: Page) => {
  await gotoURL(page, "/ui");
  await page.getByRole("link", { name: "Settings" }).click();
  await page.waitForSelector("text=Settings");
};

export const validateSettingValue = async (
  page: Page,
  setting: string,
  value: string,
) => {
  const text = setting + value;
  const settingRow = page.getByRole("row").filter({ hasText: text });
  await expect(settingRow).toBeVisible();
};

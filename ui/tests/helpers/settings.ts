import type { Page } from "@playwright/test";
import { expect } from "../fixtures/test-extension";

export const goToSettingsPage = async (page: Page) => {
  await page.goto("/ui");
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

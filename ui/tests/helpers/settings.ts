import { Locator, Page, expect } from "@playwright/test";
import { randomNameSuffix } from "./name";

export const randomClusterName = (): string => {
  return `playwright-token-${randomNameSuffix()}`;
};

export const goToSettingsPage = async (page: Page) => {
  await page.goto("/ui");
  await page.getByRole("link", { name: "Settings" }).click();
  await page.waitForSelector("text=Settings");
};

export const validateSettingValue = async (
  settingRow: Locator,
  value: string,
) => {
  const readModeValue = settingRow.locator(".readmode-value");
  await expect(readModeValue).toHaveText(value);
};

export const updateManagerSetting = async (
  page: Page,
  settingName: string,
  content: string,
): Promise<Locator> => {
  const settingRow = page.locator("css=tr", { hasText: settingName });
  await settingRow.getByRole("button").click();

  const settingInput = settingRow.getByRole("textbox");
  await settingInput.fill(content);
  await settingRow.getByRole("button", { name: "Save", exact: true }).click();
  await page.waitForSelector(`text=Setting ${settingName} updated.`);
  await page.getByRole("button", { name: "Close notification" }).click();
  return settingRow;
};

export const updateMemberSetting = async (
  page: Page,
  member: string,
  settingName: string,
  content: string,
): Promise<Locator> => {
  const memberSettingRow = page
    .getByRole("row")
    .filter({ hasText: `${member}.${settingName}` });
  await memberSettingRow.getByRole("button").click();

  const settingInput = memberSettingRow.getByRole("textbox");
  await settingInput.fill(content);
  await memberSettingRow
    .getByRole("button", { name: "Save", exact: true })
    .click();
  await page.waitForSelector(`text=Setting ${settingName} updated.`);
  await page.getByRole("button", { name: "Close notification" }).click();

  return memberSettingRow;
};

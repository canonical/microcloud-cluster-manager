import type { Page } from "@playwright/test";
import { randomNameSuffix } from "./name";
import { expect } from "../fixtures/test-extension";
import { gotoURL } from "./navigate";

export const randomClusterName = (): string => {
  return `playwright-token-${randomNameSuffix()}`;
};

export const randomSeedClusterName = (): string => {
  const randomInt = Math.floor(Math.random() * 20) + 1;
  return `cluster-${randomInt.toString().padStart(2, "0")}`;
};

export const ensureClusterExists = async (
  page: Page,
  cluster: string,
): Promise<void> => {
  await gotoURL(page, "/ui");
  const clusterNameCell = page.getByRole("row").filter({ hasText: cluster });
  await expect(clusterNameCell).toBeVisible();
};

export const goToClusterDetail = async (
  page: Page,
  cluster: string,
): Promise<void> => {
  await gotoURL(page, "/ui");
  await page.getByRole("link", { name: cluster }).click();
};

export const setClusterConfig = async (
  page: Page,
  cluster: string,
  configLabel: string,
  value: string,
): Promise<void> => {
  await goToClusterDetail(page, cluster);
  await page.getByRole("button", { name: "Configure", exact: true }).click();
  await page.getByLabel(configLabel).fill(value);
  await page.getByRole("button", { name: "Save changes" }).click();
  await expect(page.getByText(`Updated cluster ${cluster}`)).toBeVisible();
};

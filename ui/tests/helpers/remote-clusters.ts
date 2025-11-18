import type { Page } from "@playwright/test";
import { randomNameSuffix } from "./name";
import { expect } from "../fixtures/test-extension";

export const randomClusterName = (): string => {
  return `playwright-token-${randomNameSuffix()}`;
};

export const ensureClusterExists = async (
  page: Page,
  cluster: string,
): Promise<void> => {
  await page.goto("/ui");
  const clusterNameCell = page.getByRole("row").filter({ hasText: cluster });
  await expect(clusterNameCell).toBeVisible();
};

export const setClusterDiskLimit = async (
  page: Page,
  cluster: string,
  limit: string,
): Promise<void> => {
  await page.goto("/ui");
  await page.getByRole("link", { name: cluster }).click();
  await page.getByRole("button", { name: "Configure", exact: true }).click();
  await page.getByLabel("Disk threshold").fill(limit);
  await page.getByRole("button", { name: "Save changes" }).click();
  await expect(page.getByText(`Updated cluster ${cluster}`)).toBeVisible();
};

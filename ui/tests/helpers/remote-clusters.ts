import type { Page } from "@playwright/test";
import { randomNameSuffix } from "./name";
import { expect } from "../fixtures/test-extension";

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
  await page.goto("/ui");
  const clusterNameCell = page.getByRole("row").filter({ hasText: cluster });
  await expect(clusterNameCell).toBeVisible();
};

export const goToClusterDetail = async (
  page: Page,
  cluster: string,
): Promise<void> => {
  await page.goto("/ui");
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

// Real MicroClouds, see "MicroClouds for the e2e tests" in CONTRIBUTING.md.
export const tunnelCluster = "mcm-e2e-01";
export const joinCluster = "mcm-e2e-join-01";

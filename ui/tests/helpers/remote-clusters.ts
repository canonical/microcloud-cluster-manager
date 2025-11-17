import type { Page } from "@playwright/test";
import { randomNameSuffix } from "./name";
import { expect } from "../fixtures/test-extension";

export const randomClusterName = (): string => {
  return `playwright-token-${randomNameSuffix()}`;
};

export const ensureClusterExists = async (
  page: Page,
  clusterName: string,
): Promise<void> => {
  await page.goto("/ui");

  const clusterNameCell = page
    .getByRole("row")
    .filter({ hasText: clusterName });

  await expect(clusterNameCell).toBeVisible();
};

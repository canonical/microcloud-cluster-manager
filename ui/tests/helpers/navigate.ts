import type { Page } from "@playwright/test";

export const gotoURL = async (page: Page, url: string): Promise<void> => {
  await page.goto(url, { waitUntil: "commit" });
};

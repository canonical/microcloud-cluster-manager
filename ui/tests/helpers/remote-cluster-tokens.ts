import { Page, expect } from "@playwright/test";

export const createRemoteClusterToken = async (
  page: Page,
  clusterName: string,
) => {
  await page.goto("/ui");
  await page.getByRole("button", { name: "Enrol cluster" }).click();
  await page.getByPlaceholder("Enter Name").click();
  await page.getByPlaceholder("Enter Name").fill(clusterName);
  await page.getByRole("button", { name: "Create" }).click();
  await expect(
    page.getByText(
      "To finish the enrollment, run the command below on any member of the MicroCloud.",
    ),
  ).toBeVisible();
};

export const revokeRemoteClusterToken = async (
  page: Page,
  clusterName: string,
) => {
  await page.goto("/ui");
  await page.getByTestId("tab-link-Tokens").click();
  await page
    .getByRole("row", { name: clusterName })
    .getByRole("button")
    .click();
  await page
    .getByRole("dialog", { name: "Confirm revoke" })
    .getByRole("button", { name: "Revoke" })
    .click();
  await page.waitForSelector(`text=Revoked token ${clusterName}.`);
};

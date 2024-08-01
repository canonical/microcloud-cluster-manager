import { Page, expect } from "@playwright/test";
import { randomNameSuffix } from "./name";

export const randomClusterName = (): string => {
  return `playwright-token-${randomNameSuffix()}`;
};

export const getClusterCountByStatus = async (
  page: Page,
  status: "Online" | "Pending" | "Degraded",
) => {
  const clusterStatusPending = page
    .getByRole("listitem")
    .filter({ hasText: "Degraded" })
    .getByText(status);

  return parseInt(
    (await clusterStatusPending.textContent())?.split(" ")[0] || "",
  );
};

export const approvePendingCluster = async (page: Page): Promise<string> => {
  await page.goto("/ui");
  await page.getByTestId("tab-link-Pending").click();
  // Since we can create a new pending cluster in the UI, we can't target a specific cluster
  // this test currently relies on dummy data.
  const firstPendingCluster = page
    .getByRole("row", { name: "Approve" })
    .first();
  const clusterNameCell = firstPendingCluster.getByRole("gridcell").first();
  const clusterName = await clusterNameCell.innerText();

  const numPendingClustersBefore = await getClusterCountByStatus(
    page,
    "Pending",
  );

  await firstPendingCluster.getByRole("button", { name: "Approve" }).click();
  await page.waitForSelector(
    `text=Successfully approved cluster ${clusterName}.`,
  );

  const numPendingClustersAfter = await getClusterCountByStatus(
    page,
    "Pending",
  );

  expect(numPendingClustersAfter).toBeLessThan(numPendingClustersBefore);

  return clusterName;
};

export const checkClusterExistInTable = async (
  page: Page,
  clusterName: string,
  table: "Active" | "Pending",
): Promise<boolean> => {
  await page.goto("/ui");
  await page.getByTestId(`tab-link-${table}`).click();

  const tablePagination = page.getByLabel("Table pagination control");
  const nextPageButton = tablePagination.getByRole("button", {
    name: "Next page",
  });

  const clusterNameCell = page
    .getByRole("row", { name: clusterName })
    .getByRole("gridcell", { name: clusterName, exact: true });

  let clusterExists = await clusterNameCell.isVisible();
  if (clusterExists) {
    return true;
  }

  // iterage table pagination and try to find the cluster
  let isEndOfPages = await nextPageButton.isDisabled();
  while (!isEndOfPages) {
    await nextPageButton.click();
    clusterExists = await clusterNameCell.isVisible();
    if (clusterExists) {
      return true;
    }
    isEndOfPages = await nextPageButton.isDisabled();
  }

  return false;
};

export const deletePendingCluster = async (page: Page): Promise<string> => {
  await page.goto("/ui");
  await page.getByTestId("tab-link-Pending").click();
  // Since we can create a new pending cluster in the UI, we can't target a specific cluster
  // this test currently relies on dummy data.
  const firstPendingCluster = page
    .getByRole("row", { name: "Approve" })
    .first();
  const clusterNameCell = firstPendingCluster.getByRole("gridcell").first();
  const clusterName = await clusterNameCell.innerText();

  const numPendingClustersBefore = await getClusterCountByStatus(
    page,
    "Pending",
  );

  await firstPendingCluster.getByRole("button", { name: "Delete" }).click();
  await page
    .getByRole("dialog", { name: "Confirm delete" })
    .getByRole("button", { name: "Delete" })
    .click();

  await page.waitForSelector(
    `text=Successfully deleted cluster ${clusterName}.`,
  );

  const numPendingClustersAfter = await getClusterCountByStatus(
    page,
    "Pending",
  );

  expect(numPendingClustersAfter).toBeLessThan(numPendingClustersBefore);

  return clusterName;
};

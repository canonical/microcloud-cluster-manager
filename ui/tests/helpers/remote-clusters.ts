import { Page } from "@playwright/test";
import { randomNameSuffix } from "./name";

export const randomClusterName = (): string => {
  return `playwright-token-${randomNameSuffix()}`;
};

export const checkClusterExistInTable = async (
  page: Page,
  clusterName: string,
  table: "Active",
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

import { test, expect } from "./fixtures/test-extension";
import {
  clusterLinkNames,
  clusterLinkTable,
  createClusterLink,
  deleteSelectedClusterLinks,
  mockClusterLinks,
  selectClusterLinkPage,
  sourceGroupNames,
  visitClusterLinks,
} from "./helpers/cluster-links";

test.describe("Cluster links workflows", () => {
  const linkCount = 52;

  test.beforeEach(async ({ page }) => {
    await mockClusterLinks(page, linkCount);
  });

  test("navigate, create, sort, and browse grouped links", async ({ page }) => {
    await visitClusterLinks(page);
    const names = clusterLinkNames(page);
    await expect(names).toHaveCount(50);
    await expect(names.first()).toHaveText("link-50");

    await createClusterLink(page, "z-source", "a-source");
    await expect(
      page.getByRole("navigation", { name: "Table pagination control" }),
    ).toContainText(String(linkCount + 1));

    await page.getByRole("button", { name: "Link name", exact: true }).click();
    await expect(names.first()).toHaveText("link-00");
    await page.getByRole("button", { name: "Next page", exact: true }).click();
    await expect(names).toHaveText(["link-49", "link-50", "link-00"]);
    await expect(sourceGroupNames(page)).toHaveText(["z-source", "a-source"]);

    await page
      .getByRole("button", { name: "Collapse links on z-source" })
      .click();
    await expect(names).toHaveText(["link-00"]);
    await page
      .getByRole("button", { name: "Expand links on z-source" })
      .click();
    await expect(names).toHaveText(["link-49", "link-50", "link-00"]);

    await page
      .getByRole("button", { name: "Previous page", exact: true })
      .click();
    await page.getByRole("button", { name: "Cluster", exact: true }).click();
    await expect(sourceGroupNames(page)).toHaveText(["a-source", "z-source"]);
    await expect(names.nth(1)).toHaveText("link-50");
  });

  test("select across pages, delete selected links, and recover from a failed source", async ({
    page,
  }) => {
    await visitClusterLinks(page);
    await selectClusterLinkPage(page);
    await expect(page.getByText("50 cluster links selected.")).toBeVisible();
    await page.getByRole("button", { name: "Next page", exact: true }).click();
    const linkCheckboxes = clusterLinkTable(page).getByRole("checkbox", {
      name: /^Select link-/,
    });
    await expect(linkCheckboxes.nth(0)).not.toBeChecked();
    await expect(linkCheckboxes.nth(1)).not.toBeChecked();
    await page
      .getByRole("button", { name: `Select all ${linkCount} cluster links` })
      .click();
    await expect(linkCheckboxes.nth(0)).toBeChecked();
    await expect(linkCheckboxes.nth(1)).toBeChecked();
    await page.getByRole("button", { name: "Clear selection" }).click();
    await page
      .getByRole("button", { name: "Previous page", exact: true })
      .click();
    await selectClusterLinkPage(page);
    await deleteSelectedClusterLinks(page);
    await expect(clusterLinkNames(page)).toHaveText(["link-00", "link-00"]);
    await expect(
      page.getByText("Deleted 50 cluster links.", { exact: true }),
    ).toBeVisible();

    // Simulate one source failing while the other remains available.
    await page.route(
      "**/1.0/remote-cluster/a-source/cluster-links",
      async (route) => {
        await route.fulfill({
          status: 403,
          json: { type: "error", error: "Source unavailable", error_code: 403 },
        });
      },
    );
    await page.reload();
    await expect(
      page.getByText("Unable to load cluster links for a-source."),
    ).toBeVisible();
    await expect(clusterLinkNames(page)).toHaveText(["link-00"]);

    await page.unroute("**/1.0/remote-cluster/a-source/cluster-links");
    await page.reload();
    await expect(
      page.getByText("Unable to load cluster links for a-source."),
    ).toBeHidden();
    await expect(clusterLinkNames(page)).toHaveText(["link-00", "link-00"]);
  });
});

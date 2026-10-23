import { test, expect } from "./fixtures/test-extension";
import {
  clusterLinkNames,
  createClusterLink,
  mockClusterLinks,
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

    // sorting applies within each group, the group order is preserved
    await page.getByRole("button", { name: "Link name", exact: true }).click();
    await expect(names.first()).toHaveText("link-01");
    await page.getByRole("button", { name: "Link name", exact: true }).click();
    await expect(names.first()).toHaveText("link-50");

    await page.getByRole("button", { name: "Next page", exact: true }).click();
    await expect(sourceGroupNames(page)).toHaveText(["z-source", "a-source"]);

    // sorting by cluster reorders the groups on the page
    await page.getByRole("button", { name: "Cluster", exact: true }).click();
    await expect(sourceGroupNames(page)).toHaveText(["a-source", "z-source"]);

    await page.getByRole("button", { name: "Cluster", exact: true }).click();
    await expect(sourceGroupNames(page)).toHaveText(["z-source", "a-source"]);

    await page
      .getByRole("button", { name: "Collapse links on z-source" })
      .click();
    await expect(names).toHaveText(["link-00"]);
    await page
      .getByRole("button", { name: "Expand links on z-source" })
      .click();
    await expect(names).toHaveText(["link-00", "link-00"]);
  });

  test("recover from a failed source", async ({ page }) => {
    await visitClusterLinks(page);

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
    await expect(clusterLinkNames(page)).toHaveCount(50);
    await expect(clusterLinkNames(page).first()).toHaveText("link-50");

    await page.unroute("**/1.0/remote-cluster/a-source/cluster-links");
    await page.reload();
    await expect(
      page.getByText("Unable to load cluster links for a-source."),
    ).toBeHidden();
    await expect(
      page.getByRole("navigation", { name: "Table pagination control" }),
    ).toContainText(String(linkCount));
  });
});

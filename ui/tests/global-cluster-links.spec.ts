import type { Page } from "@playwright/test";
import { expect, test } from "./fixtures/test-extension";
import {
  ensureClusterExists,
  randomSeedClusterName,
} from "./helpers/remote-clusters";

interface MockLink {
  source: string;
  name: string;
  type?: string;
}

const isAllClusterLinksUrl = (url: URL) =>
  url.pathname === "/1.0/cluster-links" &&
  url.searchParams.get("all-clusters") === "true";

const mockAllClusterLinks = async (
  page: Page,
  links: MockLink[],
  errors: { cluster: string; error: string }[] = [],
) => {
  await page.route(isAllClusterLinksUrl, async (route) =>
    route.fulfill({
      json: {
        type: "sync",
        status: "Success",
        status_code: 200,
        metadata: {
          cluster_links: links.map(({ source, name, type }) => ({
            link: { name, description: "", type: type ?? "unidirectional" },
            source_cluster: source,
            target_cluster: null,
          })),
          errors,
        },
      },
    }),
  );
};

test.describe("Global cluster links page", () => {
  test("should show the rows that loaded alongside per-cluster errors", async ({
    page,
  }) => {
    await mockAllClusterLinks(
      page,
      [{ source: "alpha", name: "alpha-link" }],
      [
        { cluster: "beta", error: "connection refused" },
        { cluster: "gamma", error: "timeout" },
      ],
    );
    await page.goto("/ui/cluster-links");

    await expect(page.getByText("Loading cluster links failed")).toBeVisible();
    await expect(
      page.getByText(
        "Unable to load cluster links for beta: connection refused",
      ),
    ).toBeVisible();
    await expect(
      page.getByText("Unable to load cluster links for gamma: timeout"),
    ).toBeVisible();
    await expect(page.getByRole("link", { name: "alpha-link" })).toBeVisible();
  });

  test("should combine text and type filters", async ({ page }) => {
    await mockAllClusterLinks(page, [
      { source: "alpha", name: "images-public", type: "public" },
      { source: "alpha", name: "images-private" },
      { source: "beta", name: "backup-public", type: "public" },
    ]);
    await page.goto("/ui/cluster-links?query=images&link-type=public");

    await expect(
      page.getByRole("link", { name: "images-public" }),
    ).toBeVisible();
    await expect(
      page.getByRole("link", { name: "images-private" }),
    ).toHaveCount(0);
    await expect(page.getByRole("link", { name: "backup-public" })).toHaveCount(
      0,
    );
    await expect(page.getByRole("link", { name: "beta" })).toHaveCount(0);
  });

  test("should sort grouped rows before pagination", async ({ page }) => {
    const alphaLinks = Array.from({ length: 12 }, (_, index) => ({
      source: "alpha",
      name: `link-${String(12 - index).padStart(2, "0")}`,
    }));
    await mockAllClusterLinks(page, [
      ...alphaLinks,
      { source: "beta", name: "beta-link" },
    ]);
    await page.goto("/ui/cluster-links");

    const table = page.getByRole("grid", { name: "Cluster links" });
    await expect(table.getByRole("link", { name: "link-12" })).toBeVisible();
    await expect(table.getByRole("link", { name: "link-01" })).toHaveCount(0);

    await table.getByRole("button", { name: "Link name" }).click();
    await expect(table.getByRole("link", { name: "link-01" })).toBeVisible();
    await expect(table.getByRole("link", { name: "link-12" })).toHaveCount(0);

    await table.getByRole("button", { name: "Cluster", exact: true }).click();
    await table.getByRole("button", { name: "Cluster", exact: true }).click();
    const firstRow = table
      .getByRole("rowgroup")
      .last()
      .getByRole("row")
      .first();
    await expect(firstRow).toHaveAccessibleName(/beta/);
    await expect(table.getByRole("link", { name: "beta-link" })).toBeVisible();
  });

  test("should navigate to the cluster detail tab of the source", async ({
    page,
  }) => {
    const cluster = randomSeedClusterName();
    await ensureClusterExists(page, cluster);
    await mockAllClusterLinks(page, [{ source: cluster, name: "seed-link" }]);
    await page.goto("/ui/cluster-links");

    await page.getByRole("link", { name: "seed-link" }).click();
    await expect(page).toHaveURL(
      new RegExp(`.*/ui/cluster/${cluster}/cluster-links$`),
    );
    await expect(
      page.getByRole("tab", { name: "Cluster links" }),
    ).toHaveAttribute("aria-selected", "true");
  });
});

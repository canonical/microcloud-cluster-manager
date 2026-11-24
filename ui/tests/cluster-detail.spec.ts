import { test } from "@playwright/test";
import type { Page } from "@playwright/test";
import {
  ensureClusterExists,
  goToClusterDetail,
  setClusterConfig,
  randomSeedClusterName,
} from "./helpers/remote-clusters";
import { expect } from "./fixtures/test-extension";

test.describe("Cluster Configuration", () => {
  test("should set disk threshold limit", async ({ page }) => {
    const cluster = randomSeedClusterName();
    await ensureClusterExists(page, cluster);
    await setClusterConfig(page, cluster, "Disk threshold percentage", "42");
  });

  test("should set memory threshold limit", async ({ page }) => {
    const cluster = randomSeedClusterName();
    await ensureClusterExists(page, cluster);
    await setClusterConfig(page, cluster, "Memory threshold percentage", "42");
  });

  test("should display warnings for low thresholds and hide them when thresholds are high", async ({
    page,
  }) => {
    const cluster = randomSeedClusterName();
    await ensureClusterExists(page, cluster);
    await setClusterConfig(page, cluster, "Memory threshold percentage", "1");
    await setClusterConfig(page, cluster, "Disk threshold percentage", "1");

    await expect(page.getByText(/\d warning?/)).toBeVisible();
    const notifications = page.locator(".p-notification--caution");
    const initialCount = await notifications.count();

    await setClusterConfig(page, cluster, "Memory threshold percentage", "100");
    await setClusterConfig(page, cluster, "Disk threshold percentage", "100");
    const finalCount = await notifications.count();
    expect(initialCount === 0 || finalCount < initialCount).toBe(true);
  });
});

test.describe("Cluster Description", () => {
  test("should add description to cluster without one", async ({ page }) => {
    const cluster = randomSeedClusterName();
    await ensureClusterExists(page, cluster);
    await goToClusterDetail(page, cluster);
    await page
      .getByRole("button", { name: "Add description", exact: true })
      .click();
    await page.getByLabel("Description").fill("New description");
    await page.getByRole("button", { name: "Save changes" }).click();
    await expect(page.getByText(`Updated cluster ${cluster}`)).toBeVisible();
  });

  test("should edit existing cluster description", async ({ page }) => {
    const cluster = randomSeedClusterName();
    const initialDescription = "Initial description";
    await setClusterConfig(page, cluster, "Description", initialDescription);
    const container = page.getByText(initialDescription, { exact: true });
    const button = container.getByRole("button");
    await button.click();
    await page.getByLabel("Description").fill("Updated description");
    await page.getByRole("button", { name: "Save changes" }).click();
    await expect(page.getByText(`Updated cluster ${cluster}`)).toBeVisible();
    await expect(
      page.getByText("Updated description", { exact: true }),
    ).toBeVisible();
  });

  test("should cancel description edit", async ({ page }) => {
    const cluster = randomSeedClusterName();
    await goToClusterDetail(page, cluster);
    await page
      .getByRole("button", { name: "Add description", exact: true })
      .click();
    await page.getByLabel("Description").fill("Should not save");
    await page.getByRole("button", { name: "Cancel" }).click();
    await expect(
      page.getByText("Should not save", { exact: true }),
    ).not.toBeVisible();
  });
});

test.describe("Cluster Monitoring", () => {
  test("should display heartbeat information", async ({ page }) => {
    const cluster = randomSeedClusterName();
    await goToClusterDetail(page, cluster);
    await expect(
      page.getByText("Since last heartbeat", { exact: true }),
    ).toBeVisible();
  });
});

test.describe("Cluster Usage Statistics", () => {
  test("should display total memory usage", async ({ page }) => {
    const cluster = randomSeedClusterName();
    await goToClusterDetail(page, cluster);
    await expect(page.getByText("Total memory", { exact: true })).toBeVisible();
  });

  test("should display total storage usage", async ({ page }) => {
    const cluster = randomSeedClusterName();
    await goToClusterDetail(page, cluster);
    await expect(page.getByText("Total storage")).toBeVisible();
  });
});

test.describe("Storage Details", () => {
  test("should display storage pool details modal", async ({ page }) => {
    const cluster = randomSeedClusterName();
    await goToClusterDetail(page, cluster);
    await page.getByText("View details", { exact: true }).click();
    await expect(
      page.getByText("Storage pool details", { exact: true }),
    ).toBeVisible();
  });

  test("should close storage details modal", async ({ page }) => {
    const cluster = randomSeedClusterName();
    await goToClusterDetail(page, cluster);
    await page.getByText("View details", { exact: true }).click();
    await page.getByRole("button", { name: "Close" }).click();
    await expect(
      page.getByText("Storage pool details", { exact: true }),
    ).not.toBeVisible();
  });
});

test.describe("Cluster Members", () => {
  test("should display cluster members graph", async ({ page }) => {
    const cluster = randomSeedClusterName();
    await goToClusterDetail(page, cluster);
    const container = page.locator(".cluster-detail-doughnut-graph", {
      has: page.getByText(/\d+\s+members$/),
    });
    await expect(container).toBeVisible();
  });
});

test.describe("Cluster Instances", () => {
  test("should display cluster instances graph", async ({ page }) => {
    const cluster = randomSeedClusterName();
    await goToClusterDetail(page, cluster);
    const container = page.locator(".cluster-detail-doughnut-graph", {
      has: page.getByText(/\d+\s+instances$/),
    });
    await expect(container).toBeVisible();
  });
});

test.describe("Navigation and UI", () => {
  test("should navigate back to clusters list", async ({ page }) => {
    const cluster = randomSeedClusterName();
    await goToClusterDetail(page, cluster);
    await page.getByRole("link", { name: "Clusters", exact: true }).click();
    await expect(page).toHaveURL(/.*\/clusters$/);
  });
});

test.describe("Cluster links tab", () => {
  const mockTunnel = async (
    page: Page,
    cluster: string,
    registered: boolean,
  ): Promise<void> => {
    await page.route(`**/1.0/remote-cluster/${cluster}`, async (route) => {
      const response = await route.fetch();
      const body = (await response.json()) as {
        metadata: { tunnel_registered: boolean };
      };
      body.metadata.tunnel_registered = registered;
      await route.fulfill({ response, json: body });
    });
  };

  // Adds a managed cluster with a registered tunnel to the cluster list, so
  // that links pointing at it can be cleaned up on the target cluster.
  const mockTargetCluster = async (page: Page, name: string): Promise<void> => {
    await page.route(
      (url) => url.pathname === "/1.0/remote-cluster",
      async (route) => {
        const response = await route.fetch();
        const body = (await response.json()) as {
          metadata: Record<string, unknown>[];
        };
        body.metadata = [
          ...body.metadata,
          { ...body.metadata[0], name, tunnel_registered: true },
        ];
        await route.fulfill({ response, json: body });
      },
    );
  };

  const isClusterLinksListUrl = (url: URL, cluster: string) =>
    url.pathname === "/1.0/cluster-links" &&
    url.searchParams.get("cluster") === cluster;

  const mockClusterLinks = async (
    page: Page,
    cluster: string,
    links: { name: string; type: string }[],
    targetCluster: string | null = null,
  ) => {
    await page.route(
      (url) => isClusterLinksListUrl(url, cluster),
      async (route) =>
        route.fulfill({
          json: {
            type: "sync",
            status: "Success",
            status_code: 200,
            metadata: {
              cluster_links: links.map((link) => ({
                link,
                source_cluster: cluster,
                target_cluster: targetCluster,
              })),
              errors: [],
            },
          },
        }),
    );
  };

  test("should navigate to the Cluster links tab and back", async ({
    page,
  }) => {
    const cluster = randomSeedClusterName();
    await goToClusterDetail(page, cluster);
    await page.getByRole("tab", { name: "Cluster links" }).click();
    await expect(page).toHaveURL(
      new RegExp(`.*/ui/cluster/${cluster}/cluster-links$`),
    );
    await expect(
      page.getByRole("tab", { name: "Cluster links" }),
    ).toHaveAttribute("aria-selected", "true");

    await page.getByRole("tab", { name: "Overview" }).click();
    await expect(page).toHaveURL(new RegExp(`.*/ui/cluster/${cluster}$`));
    await expect(page.getByRole("tab", { name: "Overview" })).toHaveAttribute(
      "aria-selected",
      "true",
    );
  });

  test("should open the Cluster links tab directly from the URL", async ({
    page,
  }) => {
    const cluster = randomSeedClusterName();
    await ensureClusterExists(page, cluster);
    await page.goto(`/ui/cluster/${cluster}/cluster-links`);
    await expect(
      page.getByRole("tab", { name: "Cluster links" }),
    ).toHaveAttribute("aria-selected", "true");
  });

  test("should show tunnel requirement when tunnel is not registered", async ({
    page,
  }) => {
    const cluster = randomSeedClusterName();
    await ensureClusterExists(page, cluster);
    await mockTunnel(page, cluster, false);
    await page.goto(`/ui/cluster/${cluster}/cluster-links`);
    await expect(
      page.getByText("Cluster links require a registered tunnel."),
    ).toBeVisible();
  });

  test("should show the empty state when there are no cluster links", async ({
    page,
  }) => {
    const cluster = randomSeedClusterName();
    await ensureClusterExists(page, cluster);
    await mockTunnel(page, cluster, true);
    await mockClusterLinks(page, cluster, []);
    await page.goto(`/ui/cluster/${cluster}/cluster-links`);
    await expect(page.getByText("No cluster links found")).toBeVisible();
    await expect(
      page.getByText("There are no cluster links on this cluster."),
    ).toBeVisible();
  });

  test("should show an error when loading cluster links fails", async ({
    page,
  }) => {
    const cluster = randomSeedClusterName();
    await ensureClusterExists(page, cluster);
    await mockTunnel(page, cluster, true);
    await page.route(
      (url) => isClusterLinksListUrl(url, cluster),
      async (route) =>
        route.fulfill({
          status: 500,
          json: { type: "error", error: "boom", error_code: 500 },
        }),
    );
    await page.goto(`/ui/cluster/${cluster}/cluster-links`);
    await expect(page.getByText("Loading cluster links failed")).toBeVisible();
    await expect(page.getByText("No cluster links found")).toHaveCount(0);
  });

  test("should delete both ends of a cluster link through the source cluster", async ({
    page,
  }) => {
    const cluster = randomSeedClusterName();
    await ensureClusterExists(page, cluster);
    await mockTunnel(page, cluster, true);
    await mockTargetCluster(page, "peer");
    await mockClusterLinks(
      page,
      cluster,
      [{ name: "to-peer", type: "bidirectional" }],
      "peer",
    );
    let deleteUrl: URL | undefined;
    await page.route(
      (url) => url.pathname === "/1.0/cluster-links/to-peer",
      async (route) => {
        if (route.request().method() !== "DELETE") {
          return route.fallback();
        }
        deleteUrl = new URL(route.request().url());
        return route.fulfill({
          json: { type: "sync", status: "Success", status_code: 200 },
        });
      },
    );

    await page.goto(`/ui/cluster/${cluster}/cluster-links`);
    await page
      .getByRole("button", { name: "Delete cluster link to-peer" })
      .click();
    await expect(
      page.getByLabel("Delete links on both source and target cluster"),
    ).toBeChecked();
    await page.getByRole("button", { name: "Delete", exact: true }).click();

    await expect(page.getByText("Deleted cluster link between")).toBeVisible();
    expect(deleteUrl?.searchParams.get("cluster")).toBe(cluster);
    expect(deleteUrl?.searchParams.get("removeReverseEdge")).toBe("true");
  });
});

test.describe("Cluster Removal", () => {
  test("should remove cluster with confirmation", async ({ page }) => {
    const cluster = randomSeedClusterName();
    await goToClusterDetail(page, cluster);
    await page.getByRole("button", { name: "Remove", exact: true }).click();
    await page
      .getByRole("button", { name: "Confirm remove", exact: true })
      .click();
    await expect(page.getByText(`Removed cluster ${cluster}`)).toBeVisible();
  });

  test("should cancel cluster removal", async ({ page }) => {
    const cluster = randomSeedClusterName();
    await goToClusterDetail(page, cluster);
    await page.getByRole("button", { name: "Remove", exact: true }).click();
    await page.getByRole("button", { name: "Cancel", exact: true }).click();
    await expect(page).toHaveURL(new RegExp(`.*/cluster/${cluster}`));
    await ensureClusterExists(page, cluster);
  });
});

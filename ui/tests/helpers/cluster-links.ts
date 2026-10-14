import type { Locator, Page } from "@playwright/test";
import type { Cluster, ClusterLink } from "../../src/types/cluster";
import { expect } from "../fixtures/test-extension";
import { gotoURL } from "./navigate";

// In-memory API for multi-source pagination, creation, and deletion workflows.
export const mockClusterLinks = async (
  page: Page,
  linkCount = 52,
): Promise<void> => {
  if (!Number.isInteger(linkCount) || linkCount < 2) {
    throw new RangeError("linkCount must be an integer of at least 2");
  }

  const clusters = ["z-source", "a-source"].map(
    (name): Cluster => ({
      name,
      cluster_uuid: `uuid-${name}`,
      lxd_url: `https://${name}:8443`,
      tunnel_registered: true,
      description: "",
      disk_threshold: 80,
      memory_threshold: 80,
      ceph_count: 0,
      ceph_statuses: [],
      cpu_total_count: 0,
      cpu_load_1: "0",
      cpu_load_5: "0",
      cpu_load_15: "0",
      created_at: "2026-01-01T00:00:00Z",
      instance_count: 0,
      instance_statuses: [],
      joined_at: "2026-01-01T00:00:00Z",
      last_status_update_at: "2026-01-01T00:00:00Z",
      member_count: 0,
      member_statuses: [],
      memory_total_amount: 0,
      memory_usage: 0,
      storage_pool_usages: [],
      cluster_certificate: "",
      status: "online",
    }),
  );
  const links: Record<string, ClusterLink[]> = {
    "z-source": Array.from({ length: linkCount - 1 }, (_, index) => ({
      name: `link-${String(linkCount - 2 - index).padStart(2, "0")}`,
      description: "Link description",
      type: "bidirectional",
      config: { "volatile.uuid": "uuid-a-source" },
    })),
    "a-source": [{ name: "link-00", description: "", type: "public" }],
  };

  await page.route("**/1.0**", async (route) => {
    const request = route.request();
    const path = new URL(request.url()).pathname;
    const match =
      /^\/1.0\/remote-cluster\/([^/]+)\/cluster-links(?:\/([^/]+))?$/.exec(
        path,
      );
    let metadata: unknown;

    if (match) {
      const source = decodeURIComponent(match[1]);
      if (!links[source]) {
        throw new Error(`Unexpected source: ${source}`);
      }
      if (request.method() === "POST") {
        links[source].push(request.postDataJSON() as ClusterLink);
      } else if (request.method() === "DELETE" && match[2]) {
        links[source] = links[source].filter(
          (link) => link.name !== decodeURIComponent(match[2]),
        );
      } else if (request.method() !== "GET") {
        throw new Error(
          `Unexpected cluster-link operation: ${request.method()} ${path}`,
        );
      }
      metadata = links[source];
    } else {
      switch (path) {
        case "/1.0":
          metadata = { trusted: true, groups: ["admins"], name: "Test user" };
          break;
        case "/1.0/configuration":
          metadata = { version: { value: "test" } };
          break;
        case "/1.0/remote-cluster":
          metadata = clusters;
          break;
        case "/1.0/remote-cluster-join-token":
          metadata = [];
          break;
        default:
          throw new Error(`Unexpected mocked API request: ${path}`);
      }
    }
    await route.fulfill({
      json: { type: "sync", status: "Success", status_code: 200, metadata },
    });
  });
};

export const visitClusterLinks = async (page: Page): Promise<void> => {
  await gotoURL(page, "/ui/clusters");
  await page.getByRole("link", { name: "Cluster links", exact: true }).click();
  await expect(page).toHaveURL(/\/ui\/cluster-links$/);
  await expect(
    page.getByRole("heading", { name: "Cluster links", exact: true }),
  ).toBeVisible();
};

export const clusterLinkTable = (page: Page): Locator => {
  return page.getByRole("grid", { name: "Cluster links", exact: true });
};

export const clusterLinkNames = (page: Page): Locator => {
  return clusterLinkTable(page)
    .getByRole("rowheader")
    .getByText(/^link-\d+$/);
};

export const sourceGroupNames = (page: Page): Locator => {
  return clusterLinkTable(page)
    .getByRole("row")
    .filter({
      has: page.getByRole("button", { name: /^(Collapse|Expand) links on / }),
    })
    .getByRole("link");
};

export const createClusterLink = async (
  page: Page,
  source: string,
  target: string,
): Promise<void> => {
  await page
    .getByRole("button", { name: "Create cluster link", exact: true })
    .click();
  await expect(
    page.getByRole("heading", { name: "Create cluster link", exact: true }),
  ).toBeVisible();
  await page
    .getByRole("combobox", { name: "Source cluster", exact: true })
    .selectOption(source);
  await page
    .getByRole("combobox", { name: "Target cluster", exact: true })
    .selectOption(target);
  await page.getByRole("button", { name: "Save changes", exact: true }).click();
  await expect(
    page.getByText("Created cluster link.", { exact: true }),
  ).toBeVisible();
  await expect(
    page.getByRole("heading", { name: "Create cluster link", exact: true }),
  ).toBeHidden();
};

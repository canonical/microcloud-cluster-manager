import type { Locator, Page } from "@playwright/test";
import type { LxdApiResponse } from "types/apiResponse";
import type { ClusterLinkList } from "types/cluster";
import { expect } from "../fixtures/test-extension";

export const visitClusterLinks = async (
  page: Page,
  cluster: string,
): Promise<void> => {
  await page.goto(`/ui/cluster/${cluster}/cluster-links`);
};

export const createClusterLink = async (
  page: Page,
  sourceCluster: string,
  targetCluster: string,
): Promise<void> => {
  await page.getByRole("button", { name: "Create cluster link" }).click();
  const source = page.getByLabel("Source", { exact: true });
  const sourceTag = await source.evaluate((el) => el.tagName);
  if (sourceTag === "SELECT") {
    await source.selectOption(sourceCluster);
  } else {
    await expect(source).toHaveValue(sourceCluster);
  }
  await page.getByLabel("Target").selectOption(targetCluster);
  await page.getByRole("button", { name: "Create", exact: true }).click();
  await expect(
    page.getByText(`Created cluster link ${targetCluster}-link`),
  ).toBeVisible();
};

// Link names are only unique per source cluster, so match on both.
const getClusterLinkRow = (
  page: Page,
  linkName: string,
  sourceCluster: string,
): Locator =>
  page.getByRole("row").filter({
    has: page
      .getByRole("link", { name: linkName, exact: true })
      .and(
        page.locator(
          `[href="/ui/cluster/${encodeURIComponent(sourceCluster)}/cluster-links"]`,
        ),
      ),
  });

export const editClusterLink = async (
  page: Page,
  linkName: string,
  sourceCluster: string,
): Promise<void> => {
  const row = getClusterLinkRow(page, linkName, sourceCluster);
  await row
    .getByRole("button", { name: `Edit cluster link ${linkName}` })
    .click();
  await page.getByLabel("Description").fill("My link description");
  await page.getByRole("button", { name: "Save" }).click();
  await expect(
    page.getByText(`Cluster link ${linkName} on ${sourceCluster} saved`),
  ).toBeVisible();
  await expect(row.getByText("My link description")).toBeVisible();
};

export const deleteClusterLinkOnSource = async (
  page: Page,
  linkName: string,
  sourceCluster: string,
): Promise<void> => {
  const row = getClusterLinkRow(page, linkName, sourceCluster);
  await row
    .getByRole("button", { name: `Delete cluster link ${linkName}` })
    .click();
  const dialog = page.getByRole("dialog", { name: "Delete cluster link" });
  await expect(
    dialog.getByText(`This will permanently delete cluster link ${linkName}.`),
  ).toBeVisible();

  // LXD can return a newly created link without volatile.uuid. Without it,
  // the API cannot verify the target cluster for cleanup, so the dialog only
  // offers source-side deletion.
  const cleanupLabel = "Delete links on both source and target cluster";
  const cleanupTarget = dialog.getByRole("checkbox", { name: cleanupLabel });
  const sourceOnly = dialog.getByText("Only this side of the link is removed.");
  await expect(cleanupTarget.or(sourceOnly)).toBeVisible();
  if (await cleanupTarget.isVisible()) {
    await expect(cleanupTarget).toBeChecked();
    await dialog.getByText(cleanupLabel, { exact: true }).click();
    await expect(cleanupTarget).not.toBeChecked();
  }

  await dialog.getByRole("button", { name: "Delete", exact: true }).click();
  await expect(
    page.getByText(`Deleted cluster link ${linkName}.`),
  ).toBeVisible();
  await expect(row).toHaveCount(0);
};

// A pair of clusters can only have one link, so leftovers from a failed run
// or a previous retry must be removed before creating a new link.
export const deleteAllClusterLinks = async (
  page: Page,
  clusters: string[],
): Promise<void> => {
  const baseUrl = "/1.0/cluster-links";
  for (const cluster of clusters) {
    const response = await page.request.get(baseUrl, { params: { cluster } });
    await expect(response).toBeOK();
    const { metadata } =
      (await response.json()) as LxdApiResponse<ClusterLinkList>;
    expect(metadata.errors).toEqual([]);

    for (const { link } of metadata.cluster_links) {
      const linkUrl = `${baseUrl}/${encodeURIComponent(link.name)}`;
      const withTargetCleanup = await page.request.delete(linkUrl, {
        params: { cluster, cleanupTarget: true },
      });
      if (!withTargetCleanup.ok()) {
        const sourceOnly = await page.request.delete(linkUrl, {
          params: { cluster },
        });
        await expect(sourceOnly).toBeOK();
      }
    }
  }
};

// The tunnel is registered some time after the cluster joined.
export const waitForTunnel = async (
  page: Page,
  cluster: string,
): Promise<void> => {
  await expect
    .poll(
      async () => {
        const response = await page.request.get(
          `/1.0/remote-cluster/${cluster}`,
        );
        const body = (await response.json()) as {
          metadata?: { tunnel_registered?: boolean };
        };
        return body.metadata?.tunnel_registered;
      },
      { timeout: 120_000 },
    )
    .toBe(true);
};

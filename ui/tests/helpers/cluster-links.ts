import type { Locator, Page } from "@playwright/test";
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

export const deleteClusterLink = async (
  page: Page,
  linkName: string,
  sourceCluster: string,
  targetCluster: string,
): Promise<void> => {
  await getClusterLinkRow(page, linkName, sourceCluster)
    .getByRole("button", { name: `Delete cluster link ${linkName}` })
    .click();
  await expect(
    page.getByText(`This will permanently delete cluster link ${linkName}.`),
  ).toBeVisible();
  await expect(
    page.getByText("Delete links on both source and target cluster"),
  ).toBeVisible();
  await page
    .getByRole("dialog", { name: "Delete cluster link" })
    .getByRole("button", { name: "Delete" })
    .click();
  await expect(
    page.getByText(
      `Deleted cluster link between ${sourceCluster} and ${targetCluster}.`,
    ),
  ).toBeVisible();
};

// A pair of clusters can only have one link, so leftovers from a failed run
// or a previous retry must be removed before creating a new link.
export const deleteAllClusterLinks = async (
  page: Page,
  clusters: string[],
): Promise<void> => {
  for (const cluster of clusters) {
    const baseUrl = `/1.0/remote-cluster/${encodeURIComponent(cluster)}/cluster-links`;
    const response = await page.request.get(baseUrl);
    expect(response.ok()).toBe(true);
    const body = (await response.json()) as {
      metadata?: { name: string }[];
    };
    for (const link of body.metadata ?? []) {
      const linkUrl = `${baseUrl}/${encodeURIComponent(link.name)}`;
      const withReverse = await page.request.delete(
        `${linkUrl}?removeReverseEdge=true`,
      );
      if (!withReverse.ok()) {
        // The reverse edge may already be gone, remove this side only.
        const sourceOnly = await page.request.delete(linkUrl);
        expect(sourceOnly.ok()).toBe(true);
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

import type { Page } from "@playwright/test";
import { expect } from "../fixtures/test-extension";

export const visitClusterLinks = async (
  page: Page,
  cluster: string,
): Promise<void> => {
  await page.goto(`/ui/cluster/${cluster}/cluster-links`);
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
      { timeout: 180_000 },
    )
    .toBe(true);
};

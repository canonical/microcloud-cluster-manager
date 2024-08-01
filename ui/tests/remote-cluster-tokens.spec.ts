import { test } from "@playwright/test";
import {
  createRemoteClusterToken,
  revokeRemoteClusterToken,
} from "./helpers/remote-cluster-tokens";
import { randomClusterName } from "./helpers/remote-clusters";

test("create and revoke remote cluster token", async ({ page }) => {
  const clusterName = randomClusterName();
  await createRemoteClusterToken(page, clusterName);
  await revokeRemoteClusterToken(page, clusterName);
});

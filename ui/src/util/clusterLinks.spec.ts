import { QueryClient } from "@tanstack/react-query";
import { queryKeys } from "util/queryKeys";
import type { Cluster, ClusterLink, ResolvedClusterLink } from "types/cluster";
import {
  canCleanUpTarget,
  findCluster,
  getClusterLinkId,
  getLinkAddresses,
  getLinkDisplayAddresses,
  invalidateClusterLinkLists,
  invalidateClusterLinksAndIdentities,
} from "./clusterLinks";

const makeCluster = (name: string, tunnel_registered = true): Cluster =>
  ({ name, tunnel_registered }) as Cluster;

const makeLink = (
  name: string,
  overrides: Partial<ClusterLink> = {},
): ClusterLink => ({
  name,
  description: "",
  type: "bidirectional",
  ...overrides,
});

const makeResolvedLink = (
  link: ClusterLink,
  source_cluster = "source",
  target_cluster: string | null = null,
): ResolvedClusterLink => ({ link, source_cluster, target_cluster });

describe("cluster link query invalidation", () => {
  it("refreshes only the attempted clusters' link lists after an edit", async () => {
    const queryClient = new QueryClient();
    const invalidate = vi.spyOn(queryClient, "invalidateQueries");

    await invalidateClusterLinkLists(queryClient, []);
    expect(invalidate).not.toHaveBeenCalled();

    await invalidateClusterLinkLists(queryClient, ["source", "target"]);
    expect(invalidate.mock.calls).toEqual([
      [{ queryKey: [queryKeys.clusters, "source", queryKeys.links] }],
      [{ queryKey: [queryKeys.clusters, "target", queryKeys.links] }],
    ]);
  });

  it("refreshes link lists and identities for both ends of a bidirectional change", async () => {
    const queryClient = new QueryClient();
    const invalidate = vi.spyOn(queryClient, "invalidateQueries");

    await invalidateClusterLinksAndIdentities(
      queryClient,
      "source",
      makeLink("link"),
      "reverse",
    );

    expect(invalidate.mock.calls).toEqual([
      [{ queryKey: [queryKeys.clusters, "source", queryKeys.links] }],
      [{ queryKey: [queryKeys.clusters, "reverse", queryKeys.links] }],
      [{ queryKey: [queryKeys.clusters, "source", queryKeys.identities] }],
      [{ queryKey: [queryKeys.clusters, "reverse", queryKeys.identities] }],
    ]);
  });
});

describe("getLinkAddresses", () => {
  it("splits and trims addresses", () => {
    const link = makeLink("a", {
      config: { "volatile.addresses": "10.0.0.1:8443, 10.0.0.2:8443" },
    });
    expect(getLinkAddresses(link)).toStrictEqual([
      "10.0.0.1:8443",
      "10.0.0.2:8443",
    ]);
  });

  it("returns an empty list without config", () => {
    expect(getLinkAddresses(makeLink("a"))).toStrictEqual([]);
  });
});

describe("getLinkDisplayAddresses", () => {
  it.each([undefined, "", "  ,  "])(
    "falls back to the pending address when confirmed addresses are %s",
    (addresses) => {
      const link = makeLink("public-images", {
        type: "public",
        config: {
          "volatile.addresses": addresses,
          "volatile.pending_address": " images.example.com:8443 ",
        },
      });
      expect(getLinkDisplayAddresses(link)).toStrictEqual([
        "images.example.com:8443",
      ]);
    },
  );

  it("prefers confirmed addresses over a pending address", () => {
    const link = makeLink("public-images", {
      type: "public",
      config: {
        "volatile.addresses": "images.example.com:8443",
        "volatile.pending_address": "old.example.com:8443",
      },
    });
    expect(getLinkDisplayAddresses(link)).toStrictEqual([
      "images.example.com:8443",
    ]);
  });
});

describe("getClusterLinkId", () => {
  it("identifies a link by its source cluster and name", () => {
    expect(getClusterLinkId(makeResolvedLink(makeLink("to-two"), "one"))).toBe(
      "one/to-two",
    );
  });
});

describe("canCleanUpTarget", () => {
  it("requires a link with target cleanup and a target with a tunnel", () => {
    const link = makeLink("a");
    expect(canCleanUpTarget(link, makeCluster("two"))).toBe(true);
    expect(canCleanUpTarget(link, makeCluster("two", false))).toBe(false);
    expect(canCleanUpTarget(link, undefined)).toBe(false);
    expect(
      canCleanUpTarget(makeLink("p", { type: "public" }), makeCluster("two")),
    ).toBe(false);
  });
});

describe("findCluster", () => {
  it("finds a cluster by name and ignores unmanaged targets", () => {
    const clusters = [makeCluster("one"), makeCluster("two")];
    expect(findCluster(clusters, "two")).toBe(clusters[1]);
    expect(findCluster(clusters, "three")).toBeUndefined();
    expect(findCluster(clusters, null)).toBeUndefined();
  });
});

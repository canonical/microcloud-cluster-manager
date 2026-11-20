import type { Cluster, ClusterLink, ResolvedClusterLink } from "types/cluster";
import {
  canCleanUpTarget,
  filterClusterLinks,
  findCluster,
  getClusterLinkId,
  getClusterLinkSortData,
  getLinkAddresses,
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
      expect(getLinkAddresses(link)).toStrictEqual(["images.example.com:8443"]);
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
    expect(getLinkAddresses(link)).toStrictEqual(["images.example.com:8443"]);
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

describe("getClusterLinkSortData", () => {
  it("extracts the displayed column values for sorting", () => {
    const resolvedLink = makeResolvedLink(
      makeLink("Zulu", {
        type: "unidirectional",
        description: "Some description",
        used_by: ["/a", "/b"],
        config: { "volatile.addresses": "Z.example.com:8443" },
      }),
    );

    expect(getClusterLinkSortData(resolvedLink)).toStrictEqual({
      name: "zulu",
      type: "unidirectional",
      target: "",
      description: "some description",
      usedBy: 2,
      addresses: "z.example.com:8443",
    });
  });
});

describe("filterClusterLinks", () => {
  const links = [
    makeResolvedLink(
      makeLink("First", {
        description: "Example description",
        config: {
          "volatile.uuid": "uuid-known",
          "volatile.addresses": "10.0.0.1:8443, 10.0.0.2:8443",
        },
      }),
      "source",
      "known",
    ),
    makeResolvedLink(makeLink("other", { type: "public" })),
  ];

  it.each(["FIRST", "DESCRIPTION", "KNOWN", "10.0.0.2"])(
    "matches %s case-insensitively",
    (query) => {
      expect(
        filterClusterLinks(links, { queries: [query], types: [] }),
      ).toEqual([links[0]]);
    },
  );

  it("matches all links from a source cluster case-insensitively", () => {
    expect(
      filterClusterLinks(links, { queries: ["SOURCE"], types: [] }),
    ).toEqual(links);
  });

  it("uses OR between queries and AND between query and type filters", () => {
    expect(
      filterClusterLinks(links, {
        queries: ["first", "other"],
        types: ["public"],
      }),
    ).toEqual([links[1]]);
    expect(
      filterClusterLinks(links, { queries: ["first"], types: ["public"] }),
    ).toEqual([]);
  });

  it("returns all links without filters and handles missing config", () => {
    expect(filterClusterLinks(links, { queries: [], types: [] })).toEqual(
      links,
    );
    expect(
      filterClusterLinks(links, { queries: ["absent"], types: [] }),
    ).toEqual([]);
  });
});

import type { Cluster, ClusterLink } from "types/cluster";
import type { GlobalClusterLinkRowSortData } from "./clusterLinks";
import {
  compareGlobalClusterLinkRows,
  filterClusterLinks,
  getClusterConnections,
  getGlobalClusterLinkGroupSortData,
  getClusterLinkId,
  getGlobalClusterLinkSortData,
  getLinkAddresses,
  getLinkedCluster,
  getReverseLink,
  groupGlobalClusterLinks,
  sortGlobalClusterConnections,
} from "./clusterLinks";

const makeCluster = (name: string, uuid = "", links?: ClusterLink[]): Cluster =>
  ({
    name,
    cluster_uuid: uuid || `uuid-${name}`,
    lxd_url: `https://${name}.example.com:8443`,
    tunnel_registered: true,
    cluster_links: links,
  }) as Cluster;

const makeLink = (
  name: string,
  overrides: Partial<ClusterLink> = {},
): ClusterLink => ({
  name,
  description: "",
  type: "bidirectional",
  ...overrides,
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

describe("getClusterConnections", () => {
  it("flattens links in cluster order", () => {
    const connections = getClusterConnections([
      makeCluster("one", "", [makeLink("to-two")]),
      makeCluster("two", "", [makeLink("to-one")]),
    ]);

    expect(connections.map((connection) => connection.id)).toStrictEqual([
      getClusterLinkId("one", "to-two"),
      getClusterLinkId("two", "to-one"),
    ]);
  });

  it("skips clusters without links", () => {
    const connections = getClusterConnections([
      makeCluster("one"),
      makeCluster("two", "", [makeLink("to-one")]),
    ]);

    expect(connections).toHaveLength(1);
    expect(connections[0].source.name).toBe("two");
  });

  it("resolves targets by uuid", () => {
    const connections = getClusterConnections([
      makeCluster("one", "uuid-one", [
        makeLink("to-two", { config: { "volatile.uuid": "uuid-two" } }),
      ]),
      makeCluster("two", "uuid-two", [
        makeLink("to-one", { config: { "volatile.uuid": "uuid-one" } }),
      ]),
    ]);

    expect(connections[0].target?.name).toBe("two");
    expect(connections[1].target?.name).toBe("one");
  });

  it("resolves known targets by address when there is no UUID match", () => {
    const target = makeCluster("two");
    const connections = getClusterConnections([
      makeCluster("one", "", [
        makeLink("to-two", {
          config: { "volatile.addresses": "two.example.com:8443" },
        }),
      ]),
      target,
    ]);

    expect(connections[0].target).toBe(target);
  });

  it("leaves the target unresolved for clusters unknown to the manager", () => {
    const connections = getClusterConnections([
      makeCluster("one", "uuid-one", [
        makeLink("to-external", {
          config: {
            "volatile.uuid": "uuid-external",
            "volatile.addresses": "203.0.113.10:8443",
          },
        }),
      ]),
    ]);

    expect(connections).toHaveLength(1);
    expect(connections[0].target).toBeUndefined();
  });

  it("keeps the source cluster and link of the connections", () => {
    const source = makeCluster("one", "", [makeLink("link")]);
    const [connection] = getClusterConnections([source]);

    expect(connection).toMatchObject({
      id: getClusterLinkId("one", "link"),
      source,
      link: source.cluster_links?.[0],
    });
  });
});

describe("groupGlobalClusterLinks", () => {
  const entries = getClusterConnections([
    makeCluster("one", "", [makeLink("b-link"), makeLink("a-link")]),
    makeCluster("two", "", [makeLink("c-link")]),
  ]);

  it("groups by source cluster in cluster order and omits empty groups", () => {
    const groups = groupGlobalClusterLinks(entries.slice(0, 2));

    expect(groups.map((group) => group.cluster.name)).toStrictEqual(["one"]);
    expect(groups[0].links.map((entry) => entry.link.name)).toStrictEqual([
      "b-link",
      "a-link",
    ]);
  });
});

describe("getGlobalClusterLinkSortData", () => {
  it("extracts the displayed column values for sorting", () => {
    const entry = getClusterConnections([
      makeCluster("source", "", [
        makeLink("Zulu", {
          type: "unidirectional",
          description: "Some description",
          used_by: ["/a", "/b"],
          config: { "volatile.addresses": "Z.example.com:8443" },
        }),
      ]),
    ])[0];

    expect(getGlobalClusterLinkSortData(entry, 2)).toStrictEqual({
      group: 2,
      isGroup: 0,
      cluster: "source",
      id: getClusterLinkId("source", "Zulu"),
      name: "zulu",
      type: "unidirectional",
      target: "",
      description: "some description",
      usedBy: 2,
      addresses: "z.example.com:8443",
    });
  });
});

describe("compareGlobalClusterLinkRows", () => {
  const groupRow = (
    group: number,
    cluster: string,
  ): GlobalClusterLinkRowSortData =>
    getGlobalClusterLinkGroupSortData(cluster, group);
  const linkRow = (
    group: number,
    cluster: string,
    id: string,
    values: Record<string, unknown> = {},
  ): GlobalClusterLinkRowSortData => ({
    ...getGlobalClusterLinkGroupSortData(cluster, group),
    isGroup: 0,
    id,
    ...values,
  });

  it("sorts link rows within their group without moving rows between groups", () => {
    const rows = [
      linkRow(0, "one", "one/b", { name: "b" }),
      linkRow(0, "one", "one/a", { name: "a" }),
      linkRow(1, "two", "two/c", { name: "c" }),
    ];

    const sorted = [...rows].sort((a, b) =>
      compareGlobalClusterLinkRows(a, b, "name", "ascending"),
    );

    expect(sorted.map((row) => row.id)).toStrictEqual([
      "one/a",
      "one/b",
      "two/c",
    ]);
  });

  it("applies the sort direction within the group", () => {
    const rows = [
      linkRow(0, "one", "one/a", { name: "a" }),
      linkRow(0, "one", "one/b", { name: "b" }),
    ];

    const sorted = [...rows].sort((a, b) =>
      compareGlobalClusterLinkRows(a, b, "name", "descending"),
    );

    expect(sorted.map((row) => row.id)).toStrictEqual(["one/b", "one/a"]);
  });

  it("keeps the group header row first within its group", () => {
    const rows = [
      linkRow(0, "one", "one/a", { name: "a" }),
      groupRow(0, "one"),
    ];

    const sorted = [...rows].sort((a, b) =>
      compareGlobalClusterLinkRows(a, b, "name", "descending"),
    );

    expect(sorted.map((row) => row.id)).toStrictEqual(["group-one", "one/a"]);
  });

  it("reorders the groups when sorting by cluster and keeps link order", () => {
    const rows = [
      groupRow(0, "z-source"),
      linkRow(0, "z-source", "z-source/b", { name: "b" }),
      linkRow(0, "z-source", "z-source/a", { name: "a" }),
      groupRow(1, "a-source"),
      linkRow(1, "a-source", "a-source/c", { name: "c" }),
    ];

    const sorted = [...rows].sort((a, b) =>
      compareGlobalClusterLinkRows(a, b, "cluster", "ascending"),
    );

    expect(sorted.map((row) => row.id)).toStrictEqual([
      "group-a-source",
      "a-source/c",
      "group-z-source",
      "z-source/b",
      "z-source/a",
    ]);
  });

  it("breaks value ties by row id regardless of direction", () => {
    const rows = [
      linkRow(0, "one", "one/b", { usedBy: 1 }),
      linkRow(0, "one", "one/a", { usedBy: 1 }),
    ];

    for (const direction of ["ascending", "descending"] as const) {
      const sorted = [...rows].sort((a, b) =>
        compareGlobalClusterLinkRows(a, b, "usedBy", direction),
      );
      expect(sorted.map((row) => row.id)).toStrictEqual(["one/a", "one/b"]);
    }
  });
});

describe("filterClusterLinks", () => {
  const entries = getClusterConnections([
    makeCluster("source", "", [
      makeLink("First", {
        description: "Example description",
        config: {
          "volatile.uuid": "uuid-known",
          "volatile.addresses": "10.0.0.1:8443, 10.0.0.2:8443",
        },
      }),
      makeLink("other", { type: "public" }),
    ]),
    makeCluster("known"),
  ]);

  it.each(["FIRST", "DESCRIPTION", "KNOWN", "10.0.0.2"])(
    "matches %s case-insensitively",
    (query) => {
      expect(
        filterClusterLinks(entries, { queries: [query], types: [] }),
      ).toEqual([entries[0]]);
    },
  );

  it("uses OR between queries and AND between query and type filters", () => {
    expect(
      filterClusterLinks(entries, {
        queries: ["first", "other"],
        types: ["public"],
      }),
    ).toEqual([entries[1]]);
    expect(
      filterClusterLinks(entries, { queries: ["first"], types: ["public"] }),
    ).toEqual([]);
  });

  it("returns all entries without filters and handles missing config", () => {
    expect(filterClusterLinks(entries, { queries: [], types: [] })).toEqual(
      entries,
    );
    expect(
      filterClusterLinks(entries, { queries: ["absent"], types: [] }),
    ).toEqual([]);
  });
});

describe("getLinkedCluster", () => {
  const clusters = [
    { ...makeCluster("one"), lxd_url: "https://11.2.3.4:8443" },
    { ...makeCluster("two"), lxd_url: "https://1.2.3.4:8443" },
  ];

  it("matches an address against the host and port of the lxd URL", () => {
    const link = makeLink("a", {
      config: { "volatile.addresses": "1.2.3.4:8443" },
    });
    expect(getLinkedCluster(link, clusters)?.name).toBe("two");
  });

  it("does not match an address that is only the end of another address", () => {
    const link = makeLink("a", {
      config: { "volatile.addresses": "2.3.4:8443" },
    });
    expect(getLinkedCluster(link, clusters)).toBeUndefined();
  });

  it("matches IPv6 addresses case-insensitively", () => {
    const link = makeLink("a", {
      config: { "volatile.addresses": "[FD42::1]:8443" },
    });
    const ipv6Clusters = [
      { ...makeCluster("one"), lxd_url: "https://[fd42::1]:8443" },
    ];
    expect(getLinkedCluster(link, ipv6Clusters)?.name).toBe("one");
  });

  it("ignores clusters with an invalid lxd URL", () => {
    const link = makeLink("a", {
      config: { "volatile.addresses": "1.2.3.4:8443" },
    });
    expect(
      getLinkedCluster(link, [{ ...makeCluster("one"), lxd_url: "" }]),
    ).toBeUndefined();
  });
});

describe("getReverseLink", () => {
  const source = {
    ...makeCluster("source"),
    lxd_url: "https://1.2.3.4:8443",
  };

  it("finds the bidirectional link pointing back by uuid", () => {
    const reverse = makeLink("back", {
      config: { "volatile.uuid": source.cluster_uuid },
    });
    expect(getReverseLink([makeLink("other"), reverse], source)).toBe(reverse);
  });

  it("finds the bidirectional link pointing back by address", () => {
    const reverse = makeLink("back", {
      config: { "volatile.addresses": "1.2.3.4:8443" },
    });
    expect(getReverseLink([reverse], source)).toBe(reverse);
  });

  it("ignores links of other types that point at the cluster", () => {
    const links = [
      makeLink("public", {
        type: "public",
        config: { "volatile.uuid": source.cluster_uuid },
      }),
      makeLink("unidirectional", {
        type: "unidirectional",
        config: { "volatile.addresses": "1.2.3.4:8443" },
      }),
    ];
    expect(getReverseLink(links, source)).toBeUndefined();
  });

  it("does not match an address that is only the end of another address", () => {
    const reverse = makeLink("back", {
      config: { "volatile.addresses": "2.3.4:8443" },
    });
    expect(getReverseLink([reverse], source)).toBeUndefined();
  });
});

describe("sortGlobalClusterConnections", () => {
  const connections = getClusterConnections([
    makeCluster("one", "", [makeLink("b"), makeLink("a")]),
    makeCluster("two", "", [makeLink("d"), makeLink("c")]),
  ]);
  const ids = (sorted: typeof connections) =>
    sorted.map((connection) => connection.id);

  it("keeps the order without a sort key", () => {
    expect(sortGlobalClusterConnections(connections, null, "ascending")).toBe(
      connections,
    );
  });

  it("sorts links within their cluster and keeps the cluster order", () => {
    expect(
      ids(sortGlobalClusterConnections(connections, "name", "ascending")),
    ).toStrictEqual(["one/a", "one/b", "two/c", "two/d"]);
    expect(
      ids(sortGlobalClusterConnections(connections, "name", "descending")),
    ).toStrictEqual(["one/b", "one/a", "two/d", "two/c"]);
  });

  it("reorders the clusters when sorting by cluster", () => {
    expect(
      ids(sortGlobalClusterConnections(connections, "cluster", "descending")),
    ).toStrictEqual(["two/d", "two/c", "one/b", "one/a"]);
  });

  it("does not change the given connections", () => {
    const before = ids(connections);
    sortGlobalClusterConnections(connections, "name", "ascending");
    expect(ids(connections)).toStrictEqual(before);
  });
});

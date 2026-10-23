import type { Cluster, ClusterLink } from "types/cluster";
import type { ClusterLinkRowSortData } from "./clusterLinks";
import {
  compareClusterLinkRows,
  flattenClusterLinks,
  filterClusterLinks,
  getClusterLinkGroupSortData,
  getClusterLinkId,
  getClusterLinkSortData,
  getLinkAddresses,
  groupClusterLinks,
} from "./clusterLinks";

const makeCluster = (name: string, uuid = ""): Cluster =>
  ({
    name,
    cluster_uuid: uuid || `uuid-${name}`,
    lxd_url: `https://${name}.example.com:8443`,
    tunnel_registered: true,
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

describe("flattenClusterLinks", () => {
  it("flattens links in cluster order and resolves targets by uuid", () => {
    const clusters = [
      makeCluster("one", "uuid-one"),
      makeCluster("two", "uuid-two"),
    ];
    const linksByCluster = {
      two: [makeLink("to-one", { config: { "volatile.uuid": "uuid-one" } })],
      one: [makeLink("to-two", { config: { "volatile.uuid": "uuid-two" } })],
    };

    const entries = flattenClusterLinks(clusters, linksByCluster);

    expect(entries.map((entry) => entry.id)).toStrictEqual([
      getClusterLinkId("one", "to-two"),
      getClusterLinkId("two", "to-one"),
    ]);
    expect(entries[0].target?.name).toBe("two");
    expect(entries[1].target?.name).toBe("one");
  });

  it("skips clusters without link results", () => {
    const clusters = [makeCluster("one"), makeCluster("two")];
    const entries = flattenClusterLinks(clusters, {
      two: [makeLink("to-one")],
    });

    expect(entries).toHaveLength(1);
    expect(entries[0].source.name).toBe("two");
    expect(entries[0].target).toBeUndefined();
  });

  it("resolves targets outside the query sources without including their links", () => {
    const source = makeCluster("one");
    const target = makeCluster("two");
    const entries = flattenClusterLinks(
      [source],
      {
        one: [
          makeLink("to-two", {
            config: { "volatile.uuid": target.cluster_uuid },
          }),
        ],
        two: [makeLink("not-queried")],
      },
      [source, target],
    );

    expect(entries).toHaveLength(1);
    expect(entries[0].source).toBe(source);
    expect(entries[0].target).toBe(target);
  });

  it("resolves known targets by address when there is no UUID match", () => {
    const source = makeCluster("one");
    const target = makeCluster("two");
    const entries = flattenClusterLinks(
      [source],
      {
        one: [
          makeLink("to-two", {
            config: { "volatile.addresses": "two.example.com:8443" },
          }),
        ],
      },
      [source, target],
    );
    expect(entries[0].target).toBe(target);
  });

  it("leaves the target unresolved for clusters unknown to the manager", () => {
    const clusters = [makeCluster("one", "uuid-one")];
    const entries = flattenClusterLinks(clusters, {
      one: [
        makeLink("to-external", {
          config: {
            "volatile.uuid": "uuid-external",
            "volatile.addresses": "203.0.113.10:8443",
          },
        }),
      ],
    });

    expect(entries).toHaveLength(1);
    expect(entries[0].target).toBeUndefined();
  });
});

describe("groupClusterLinks", () => {
  const clusters = [makeCluster("one"), makeCluster("two")];
  const entries = flattenClusterLinks(clusters, {
    one: [makeLink("b-link"), makeLink("a-link")],
    two: [makeLink("c-link")],
  });

  it("groups by source cluster in cluster order and omits empty groups", () => {
    const groups = groupClusterLinks(entries.slice(0, 2));

    expect(groups.map((group) => group.cluster.name)).toStrictEqual(["one"]);
    expect(groups[0].links.map((entry) => entry.link.name)).toStrictEqual([
      "b-link",
      "a-link",
    ]);
  });
});

describe("getClusterLinkSortData", () => {
  it("extracts the displayed column values for sorting", () => {
    const source = makeCluster("source");
    const entry = flattenClusterLinks([source], {
      source: [
        makeLink("Zulu", {
          type: "unidirectional",
          description: "Some description",
          used_by: ["/a", "/b"],
          config: { "volatile.addresses": "Z.example.com:8443" },
        }),
      ],
    })[0];

    expect(getClusterLinkSortData(entry, 2)).toStrictEqual({
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

describe("compareClusterLinkRows", () => {
  const groupRow = (group: number, cluster: string): ClusterLinkRowSortData =>
    getClusterLinkGroupSortData(cluster, group);
  const linkRow = (
    group: number,
    cluster: string,
    id: string,
    values: Record<string, unknown> = {},
  ): ClusterLinkRowSortData => ({
    ...getClusterLinkGroupSortData(cluster, group),
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
      compareClusterLinkRows(a, b, "name", "ascending"),
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
      compareClusterLinkRows(a, b, "name", "descending"),
    );

    expect(sorted.map((row) => row.id)).toStrictEqual(["one/b", "one/a"]);
  });

  it("keeps the group header row first within its group", () => {
    const rows = [
      linkRow(0, "one", "one/a", { name: "a" }),
      groupRow(0, "one"),
    ];

    const sorted = [...rows].sort((a, b) =>
      compareClusterLinkRows(a, b, "name", "descending"),
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
      compareClusterLinkRows(a, b, "cluster", "ascending"),
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
        compareClusterLinkRows(a, b, "usedBy", direction),
      );
      expect(sorted.map((row) => row.id)).toStrictEqual(["one/a", "one/b"]);
    }
  });
});

describe("filterClusterLinks", () => {
  const source = makeCluster("source");
  const entries = flattenClusterLinks(
    [source],
    {
      source: [
        makeLink("First", {
          description: "Example description",
          config: {
            "volatile.uuid": "uuid-known",
            "volatile.addresses": "10.0.0.1:8443, 10.0.0.2:8443",
          },
        }),
        makeLink("other", { type: "public" }),
      ],
    },
    [source, makeCluster("known")],
  );

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

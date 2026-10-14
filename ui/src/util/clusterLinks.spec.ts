import type { Cluster, ClusterLink } from "types/cluster";
import {
  flattenClusterLinks,
  filterClusterLinks,
  getClusterLinkId,
  getLinkAddresses,
  groupClusterLinks,
  sortClusterLinks,
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

  it("sorts within groups without moving rows between groups", () => {
    const groups = groupClusterLinks(
      sortClusterLinks(entries, "name", "ascending", true),
    );

    expect(groups.map((group) => group.cluster.name)).toStrictEqual([
      "one",
      "two",
    ]);
    expect(groups[0].links.map((entry) => entry.link.name)).toStrictEqual([
      "a-link",
      "b-link",
    ]);
    expect(groups[1].links.map((entry) => entry.link.name)).toStrictEqual([
      "c-link",
    ]);
  });

  it("sorts the groups themselves when sorting by cluster", () => {
    const groups = groupClusterLinks(
      sortClusterLinks(entries, "cluster", "descending", true),
    );

    expect(groups.map((group) => group.cluster.name)).toStrictEqual([
      "two",
      "one",
    ]);
    // link order within groups is untouched
    expect(groups[1].links.map((entry) => entry.link.name)).toStrictEqual([
      "b-link",
      "a-link",
    ]);
  });
});

describe("sortClusterLinks", () => {
  const source = makeCluster("source");
  const entries = flattenClusterLinks(
    [source],
    {
      source: [
        makeLink("Zulu", {
          type: "unidirectional",
          description: "Zulu description",
          used_by: ["/a", "/b"],
          config: {
            "volatile.uuid": "uuid-z-target",
            "volatile.addresses": "z.example.com:8443",
          },
        }),
        makeLink("alpha", {
          type: "bidirectional",
          description: "Alpha description",
          used_by: ["/a"],
          config: {
            "volatile.uuid": "uuid-a-target",
            "volatile.addresses": "a.example.com:8443",
          },
        }),
      ],
    },
    [source, makeCluster("a-target"), makeCluster("z-target")],
  );

  it.each([
    "name",
    "type",
    "target",
    "description",
    "usedBy",
    "addresses",
  ] as const)("sorts %s in both directions without mutating input", (key) => {
    expect(
      sortClusterLinks(entries, key).map((entry) => entry.link.name),
    ).toEqual(["alpha", "Zulu"]);
    expect(
      sortClusterLinks(entries, key, "descending").map(
        (entry) => entry.link.name,
      ),
    ).toEqual(["Zulu", "alpha"]);
    expect(entries.map((entry) => entry.link.name)).toEqual(["Zulu", "alpha"]);
  });

  it("retains the original order when sorting is cleared", () => {
    expect(sortClusterLinks(entries, null)).toEqual(entries);
    expect(sortClusterLinks(entries, null)).not.toBe(entries);
  });

  it.each(["target", "description", "usedBy", "addresses"] as const)(
    "handles missing %s values and breaks ties by ID",
    (key) => {
      const missing = flattenClusterLinks([source], {
        source: [makeLink("b"), makeLink("a")],
      });
      expect(
        sortClusterLinks(missing, key).map((entry) => entry.link.name),
      ).toEqual(["a", "b"]);
      expect(
        sortClusterLinks(missing, key, "descending").map(
          (entry) => entry.link.name,
        ),
      ).toEqual(["a", "b"]);
    },
  );
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

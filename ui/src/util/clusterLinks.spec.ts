import type { Cluster, ClusterLink } from "types/cluster";
import {
  filterClusterLinks,
  getClusterLinkId,
  getClusterLinkSortData,
  getLinkAddresses,
  getLinkedCluster,
  getReverseLink,
  getSourceClusterConnections,
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

describe("getSourceClusterConnections", () => {
  it("resolves targets by uuid", () => {
    const one = makeCluster("one", "uuid-one");
    const two = makeCluster("two", "uuid-two");
    const connections = getSourceClusterConnections(
      one,
      [makeLink("to-two", { config: { "volatile.uuid": "uuid-two" } })],
      [one, two],
    );

    expect(connections[0].target).toBe(two);
  });

  it("resolves known targets by address when there is no UUID match", () => {
    const target = makeCluster("two");
    const connections = getSourceClusterConnections(
      makeCluster("one"),
      [
        makeLink("to-two", {
          config: { "volatile.addresses": "two.example.com:8443" },
        }),
      ],
      [target],
    );

    expect(connections[0].target).toBe(target);
  });

  it("resolves a pending public link to a registered target by its pending address", () => {
    const target = makeCluster("two");
    const connections = getSourceClusterConnections(
      makeCluster("one"),
      [
        makeLink("public-images", {
          type: "public",
          config: { "volatile.pending_address": "two.example.com:8443" },
        }),
      ],
      [target],
    );

    expect(connections[0].target).toBe(target);
  });

  it("leaves the target unresolved for clusters unknown to the manager", () => {
    const source = makeCluster("one", "uuid-one");
    const connections = getSourceClusterConnections(
      source,
      [
        makeLink("to-external", {
          config: {
            "volatile.uuid": "uuid-external",
            "volatile.addresses": "203.0.113.10:8443",
          },
        }),
      ],
      [source],
    );

    expect(connections).toHaveLength(1);
    expect(connections[0].target).toBeUndefined();
  });

  it("keeps the source cluster and link of the connections", () => {
    const source = makeCluster("one");
    const link = makeLink("link");
    const [connection] = getSourceClusterConnections(source, [link], [source]);

    expect(connection).toMatchObject({
      id: getClusterLinkId("one", "link"),
      source,
      link,
    });
  });
});

describe("getClusterLinkSortData", () => {
  it("extracts the displayed column values for sorting", () => {
    const source = makeCluster("source");
    const [entry] = getSourceClusterConnections(
      source,
      [
        makeLink("Zulu", {
          type: "unidirectional",
          description: "Some description",
          used_by: ["/a", "/b"],
          config: { "volatile.addresses": "Z.example.com:8443" },
        }),
      ],
      [source],
    );

    expect(getClusterLinkSortData(entry)).toStrictEqual({
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
  const entries = getSourceClusterConnections(
    makeCluster("source"),
    [
      makeLink("First", {
        description: "Example description",
        config: {
          "volatile.uuid": "uuid-known",
          "volatile.addresses": "10.0.0.1:8443, 10.0.0.2:8443",
        },
      }),
      makeLink("other", { type: "public" }),
    ],
    [makeCluster("known")],
  );

  it.each(["FIRST", "DESCRIPTION", "KNOWN", "10.0.0.2"])(
    "matches %s case-insensitively",
    (query) => {
      expect(
        filterClusterLinks(entries, { queries: [query], types: [] }),
      ).toEqual([entries[0]]);
    },
  );

  it("matches all links from a source cluster case-insensitively", () => {
    expect(
      filterClusterLinks(entries, { queries: ["SOURCE"], types: [] }),
    ).toEqual(entries);
  });

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

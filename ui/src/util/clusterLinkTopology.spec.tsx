import type { Cluster, ClusterLink, ClusterLinkType } from "types/cluster";
import { toClusterLinkTopologyItem } from "./clusterLinkTopology";
import { act } from "react";
import { createRoot } from "react-dom/client";
import { MemoryRouter } from "react-router-dom";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import ClusterTopology from "pages/clusters/ClusterTopology";
import { fetchClusterLinks, fetchClusters } from "api/clusters";

vi.mock("api/clusters", () => ({
  fetchClusterLinks: vi.fn(),
  fetchClusters: vi.fn(),
}));

vi.mock("components/DsIcon", () => ({
  default: () => <svg aria-hidden="true" />,
}));

const targetCluster: Cluster = {
  name: "target",
  cluster_uuid: "target-uuid",
  lxd_url: "https://target.example.test:8443",
  description: "",
  disk_threshold: 0,
  memory_threshold: 0,
  ceph_count: 0,
  ceph_statuses: [],
  cpu_total_count: 0,
  cpu_load_1: "",
  cpu_load_5: "",
  cpu_load_15: "",
  created_at: "",
  instance_count: 0,
  instance_statuses: [],
  joined_at: "",
  last_status_update_at: "",
  member_count: 0,
  member_statuses: [],
  memory_total_amount: 0,
  memory_usage: 0,
  storage_pool_usages: [],
  cluster_certificate: "",
  status: "online",
  tunnel_registered: true,
};

const continuationCases = [
  {
    name: "enrolled target with no onward links",
    enrolled: true,
    onward: false,
    dots: 0,
  },
  {
    name: "enrolled target with onward links",
    enrolled: true,
    onward: true,
    dots: 1,
  },
  {
    name: "enrolled target with missing enrichment",
    enrolled: true,
    onward: undefined,
    dots: 1,
  },
  {
    name: "unknown target with no onward links",
    enrolled: false,
    onward: false,
    dots: 1,
  },
  {
    name: "unknown target with onward links",
    enrolled: false,
    onward: true,
    dots: 1,
  },
  {
    name: "unknown target with missing enrichment",
    enrolled: false,
    onward: undefined,
    dots: 1,
  },
];

describe("toClusterLinkTopologyItem", () => {
  it("uses the link name as its stable ID and resolves the target by UUID", () => {
    const link: ClusterLink = {
      name: "custom-name",
      description: "Original link description",
      type: "bidirectional",
      config: { "volatile.uuid": targetCluster.cluster_uuid },
    };

    expect(toClusterLinkTopologyItem(link, [targetCluster])).toEqual({
      clusterLink: link,
      id: "custom-name",
      name: "custom-name",
      type: "bidirectional",
      targetClusterName: "target",
      hasMore: true,
    });
    expect(toClusterLinkTopologyItem(link, [targetCluster]).clusterLink).toBe(
      link,
    );
  });

  it.each(continuationCases)(
    "uses backend enrichment for $name",
    ({ enrolled, onward, dots }) => {
      // Arrange
      const link: ClusterLink = {
        name: "source-link",
        description: "",
        type: "bidirectional",
        config: { "volatile.uuid": targetCluster.cluster_uuid },
        ...(onward === undefined ? {} : { has_onward_cluster_links: onward }),
      };

      // Act
      const item = toClusterLinkTopologyItem(
        link,
        enrolled ? [targetCluster] : [],
      );

      // Assert
      expect(item.hasMore).toBe(dots === 1);
      expect(item.targetClusterName).toBe(
        enrolled ? targetCluster.name : undefined,
      );
      expect(item.clusterLink).toBe(link);
      expect(item.id).toBe(link.name);
    },
  );

  it("falls back to target addresses without inferring names from link names", () => {
    expect(
      toClusterLinkTopologyItem(
        {
          name: "unrelated-name",
          description: "",
          type: "unidirectional",
          config: { "volatile.addresses": "target.example.test:8443" },
        },
        [targetCluster],
      ).targetClusterName,
    ).toBe("target");
  });

  it.each<["bidirectional" | "unidirectional" | "public"]>([
    ["bidirectional"],
    ["unidirectional"],
    ["public"],
  ])("preserves the %s type and shows dots for an unknown target", (type) => {
    expect(
      toClusterLinkTopologyItem(
        { name: "external-link", description: "", type },
        [],
      ),
    ).toEqual({
      clusterLink: { name: "external-link", description: "", type },
      id: "external-link",
      name: "external-link",
      type,
      targetClusterName: undefined,
      hasMore: true,
    });
  });

  it("does not mutate the API link", () => {
    const link = Object.freeze({
      name: "link",
      description: "",
      type: "public",
      has_onward_cluster_links: false,
      config: Object.freeze({ "volatile.uuid": targetCluster.cluster_uuid }),
    });
    const original = structuredClone(link);

    const item = toClusterLinkTopologyItem(link, [targetCluster]);

    expect(link).toEqual(original);
    expect(item.clusterLink).toBe(link);
  });

  it("reports unsupported API link types instead of casting them", () => {
    expect(() =>
      toClusterLinkTopologyItem(
        {
          name: "link",
          description: "",
          type: "unsupported" as ClusterLinkType,
        },
        [],
      ),
    ).toThrow('Unsupported cluster link type "unsupported" for "link".');
  });
});

describe("ClusterTopology continuation", () => {
  const sourceCluster: Cluster = {
    ...targetCluster,
    name: "source",
    cluster_uuid: "source-uuid",
    lxd_url: "https://source.example.test:8443",
  };
  const forwardLink: ClusterLink = {
    name: "abc-def-clusterlink",
    description: "",
    type: "bidirectional",
    config: { "volatile.uuid": targetCluster.cluster_uuid },
  };
  let client: QueryClient;
  let container: HTMLDivElement;
  let root: ReturnType<typeof createRoot>;

  beforeEach(() => {
    vi.mocked(fetchClusterLinks).mockReset();
    vi.mocked(fetchClusters).mockReset();
    vi.mocked(fetchClusters).mockResolvedValue([sourceCluster, targetCluster]);
    client = new QueryClient({
      defaultOptions: { queries: { retry: false } },
    });
    container = document.createElement("div");
    root = createRoot(container);
  });

  afterEach(() => {
    act(() => {
      root.unmount();
    });
    client.clear();
  });

  const renderTopology = () => {
    act(() => {
      root.render(
        <MemoryRouter>
          <QueryClientProvider client={client}>
            <ClusterTopology cluster={sourceCluster} />
          </QueryClientProvider>
        </MemoryRouter>,
      );
    });
  };

  const waitForContent = async (assertion: () => void) => {
    await vi.waitFor(async () => {
      await act(async () => {
        await new Promise((resolve) => setTimeout(resolve, 0));
      });
      assertion();
    });
  };

  it.each(continuationCases)(
    "renders $name using only enriched source links",
    async ({ onward, dots, enrolled }) => {
      // Arrange
      vi.mocked(fetchClusters).mockResolvedValue(
        enrolled ? [sourceCluster, targetCluster] : [sourceCluster],
      );
      vi.mocked(fetchClusterLinks).mockResolvedValue([
        {
          ...forwardLink,
          ...(onward === undefined ? {} : { has_onward_cluster_links: onward }),
        },
      ]);

      // Act
      renderTopology();

      // Assert
      await waitForContent(() => {
        expect(container.querySelector(".downstream-item")).not.toBeNull();
      });
      const targetChip = container.querySelector(
        ".downstream-item .target-cell .p-chip__value",
      );
      expect(targetChip?.textContent).toBe(
        enrolled ? targetCluster.name : undefined,
      );
      expect(
        container.querySelector(".link-cell .p-chip__value")?.textContent,
      ).toBe(forwardLink.name);
      expect(container.querySelectorAll(".downstream-item")).toHaveLength(1);
      expect(container.querySelectorAll(".trail-dots")).toHaveLength(dots);
      expect(container.querySelectorAll(".target-connector")).toHaveLength(
        1 + (enrolled ? dots : 0),
      );
      expect(fetchClusterLinks).toHaveBeenCalledExactlyOnceWith(
        sourceCluster.name,
        "1",
      );
    },
  );

  it("expands more than four links without duplicates or additional requests", async () => {
    // Arrange
    const links: ClusterLink[] = Array.from({ length: 6 }, (_, index) => ({
      ...forwardLink,
      name: `source-link-${index}`,
      has_onward_cluster_links: false,
    }));
    vi.mocked(fetchClusterLinks).mockResolvedValue(links);
    renderTopology();
    await waitForContent(() => {
      expect(container.querySelector("button")?.textContent).toBe(
        "Show 2 more",
      );
    });
    expect(container.querySelectorAll(".link-cell")).toHaveLength(4);
    expect(
      Array.from(
        container.querySelectorAll(".link-cell .p-chip__value"),
        (chip) => chip.textContent,
      ),
    ).toEqual(links.slice(0, 4).map((link) => link.name));

    // Act
    act(() => {
      const button = container.querySelector("button");
      if (!button) {
        throw new Error("Expected the Show more button");
      }
      button.click();
    });

    // Assert
    expect(container.querySelector("button")).toBeNull();
    expect(container.querySelectorAll(".downstream-item")).toHaveLength(
      links.length,
    );
    expect(
      Array.from(
        container.querySelectorAll(".link-cell .p-chip__value"),
        (chip) => chip.textContent,
      ),
    ).toEqual(links.map((link) => link.name));
    expect(
      container.querySelectorAll(".target-cell .p-chip__value"),
    ).toHaveLength(links.length);
    expect(container.querySelectorAll(".target-connector")).toHaveLength(
      links.length,
    );
    expect(container.querySelectorAll(".trail-dots")).toHaveLength(0);
    expect(fetchClusterLinks).toHaveBeenCalledExactlyOnceWith(
      sourceCluster.name,
      "1",
    );
  });

  it("shows loading until the enriched source links resolve", async () => {
    // Arrange
    let resolveLinks!: (links: ClusterLink[]) => void;
    vi.mocked(fetchClusterLinks).mockReturnValue(
      new Promise<ClusterLink[]>((resolve) => {
        resolveLinks = resolve;
      }),
    );
    renderTopology();
    await waitForContent(() => {
      expect(container.textContent).toContain("Loading cluster topology...");
      expect(fetchClusterLinks).toHaveBeenCalledExactlyOnceWith(
        sourceCluster.name,
        "1",
      );
    });
    expect(container.querySelector(".downstream-item")).toBeNull();

    // Act
    act(() => {
      resolveLinks([{ ...forwardLink, has_onward_cluster_links: false }]);
    });

    // Assert
    await waitForContent(() => {
      expect(container.querySelector(".downstream-item")).not.toBeNull();
    });
    expect(container.textContent).not.toContain("Loading cluster topology...");
    expect(container.querySelectorAll(".trail-dots")).toHaveLength(0);
    expect(fetchClusterLinks).toHaveBeenCalledExactlyOnceWith(
      sourceCluster.name,
      "1",
    );
  });

  it("shows an error when the source links cannot be loaded", async () => {
    // Arrange
    vi.mocked(fetchClusterLinks).mockRejectedValue(
      new Error("Source links unavailable"),
    );

    // Act
    renderTopology();

    // Assert
    await waitForContent(() => {
      expect(container.textContent).toContain(
        "Loading cluster topology failed",
      );
      expect(container.textContent).toContain("Source links unavailable");
    });
    expect(container.querySelector(".downstream-item")).toBeNull();
    expect(container.querySelectorAll(".trail-dots")).toHaveLength(0);
    expect(fetchClusterLinks).toHaveBeenCalledExactlyOnceWith(
      sourceCluster.name,
      "1",
    );
  });
});

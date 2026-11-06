import type { Cluster, ClusterLink, ClusterLinkType } from "types/cluster";
import {
  hasOnwardClusterLinks,
  toClusterLinkTopologyItem,
} from "./clusterLinkTopology";
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
      hasMore: undefined,
    });
    expect(toClusterLinkTopologyItem(link, [targetCluster]).clusterLink).toBe(
      link,
    );
  });

  describe("hasOnwardClusterLinks", () => {
    const reverseLink: ClusterLink = {
      name: "arbitrary-reverse-name",
      description: "",
      type: "bidirectional",
      config: { "volatile.uuid": targetCluster.cluster_uuid },
    };

    it("returns false for no links", () => {
      expect(hasOnwardClusterLinks([], targetCluster)).toBe(false);
    });

    it("ignores a reverse link identified by UUID, regardless of its name", () => {
      expect(hasOnwardClusterLinks([reverseLink], targetCluster)).toBe(false);
    });

    it("ignores reverse links identified by address", () => {
      expect(
        hasOnwardClusterLinks(
          [
            {
              name: "reverse",
              description: "",
              type: "unidirectional",
              config: { "volatile.addresses": "target.example.test:8443" },
            },
          ],
          targetCluster,
        ),
      ).toBe(false);
    });

    it("counts additional links even when their targets are not enrolled", () => {
      expect(
        hasOnwardClusterLinks(
          [reverseLink, { name: "external", description: "", type: "public" }],
          targetCluster,
        ),
      ).toBe(true);
    });

    it("counts an onward link even if it has the same name as the source link", () => {
      expect(
        hasOnwardClusterLinks(
          [
            {
              name: "abc-def-clusterlink",
              description: "",
              type: "bidirectional",
              config: { "volatile.uuid": "another-cluster-uuid" },
            },
          ],
          targetCluster,
        ),
      ).toBe(true);
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
    const reverseLink: ClusterLink = {
      name: "def-abc",
      description: "",
      type: "bidirectional",
      config: { "volatile.uuid": sourceCluster.cluster_uuid },
    };

    it.each([
      {
        name: "no onward links",
        targetLinks: [] as ClusterLink[],
        dots: 0,
        enrolled: true,
      },
      {
        name: "only a reverse link",
        targetLinks: [reverseLink] as ClusterLink[],
        dots: 0,
        enrolled: true,
      },
      {
        name: "an onward link in addition to the reverse link",
        targetLinks: [
          reverseLink,
          { name: "external", type: "public" } as ClusterLink,
        ],
        dots: 1,
        enrolled: true,
      },
      {
        name: "an unknown target",
        targetLinks: [] as ClusterLink[],
        dots: 1,
        enrolled: false,
      },
    ])(
      "renders the topology for $name",
      async ({ targetLinks, dots, enrolled }) => {
        vi.mocked(fetchClusterLinks).mockClear();
        vi.mocked(fetchClusters).mockResolvedValue(
          enrolled ? [sourceCluster, targetCluster] : [sourceCluster],
        );
        vi.mocked(fetchClusterLinks).mockImplementation(
          async (name) =>
            await Promise.resolve(
              name === sourceCluster.name ? [forwardLink] : targetLinks,
            ),
        );
        const client = new QueryClient({
          defaultOptions: { queries: { retry: false } },
        });
        const container = document.createElement("div");
        const root = createRoot(container);
        try {
          act(() => {
            root.render(
              <MemoryRouter>
                <QueryClientProvider client={client}>
                  <ClusterTopology cluster={sourceCluster} />
                </QueryClientProvider>
              </MemoryRouter>,
            );
          });
          await vi.waitFor(async () => {
            await act(async () => {
              await new Promise((resolve) => setTimeout(resolve, 10));
            });
            expect(container.querySelector(".downstream-item")).not.toBeNull();
          });
          expect(
            container.querySelector(".target-cluster-chip")?.textContent,
          ).toBe(enrolled ? targetCluster.name : undefined);
          expect(container.querySelectorAll(".trail-dots")).toHaveLength(dots);
          expect(container.querySelectorAll(".target-connector")).toHaveLength(
            1 + (enrolled ? dots : 0),
          );
          expect(fetchClusterLinks).toHaveBeenCalledTimes(enrolled ? 2 : 1);
        } finally {
          act(() => {
            root.unmount();
          });
          client.clear();
        }
      },
    );
  });
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
      config: Object.freeze({ "volatile.uuid": "unmanaged-uuid" }),
    });
    toClusterLinkTopologyItem(link, [targetCluster]);
    expect(link.config["volatile.uuid"]).toBe("unmanaged-uuid");
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

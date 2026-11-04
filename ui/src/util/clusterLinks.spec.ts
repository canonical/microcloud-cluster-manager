import { isPendingPublicLink } from "./clusterLinks";
import type { ClusterLink } from "types/cluster";

describe("isPendingPublicLink", () => {
  it("recognizes public links without confirmed addresses", () => {
    const link: ClusterLink = {
      name: "target",
      type: "public",
      description: "",
      config: { "volatile.pending_certificate": "certificate" },
    };

    expect(isPendingPublicLink(link)).toBe(true);
  });

  it("does not mark confirmed links or other link types as pending", () => {
    const confirmed: ClusterLink = {
      name: "target",
      type: "public",
      description: "",
      config: { "volatile.addresses": "10.0.0.2:8443" },
    };
    const bidirectional: ClusterLink = {
      name: "target",
      type: "bidirectional",
      description: "",
    };

    expect(isPendingPublicLink(confirmed)).toBe(false);
    expect(isPendingPublicLink(bidirectional)).toBe(false);
  });
});

import { getCertificateFingerprint, isPendingPublicLink } from "./clusterLinks";
import type { ClusterLink } from "types/cluster";

describe("isPendingPublicLink", () => {
  it("recognizes public links without confirmed addresses", () => {
    const link: ClusterLink = {
      name: "target",
      type: "public",
      config: { "volatile.pending_certificate": "certificate" },
    };

    expect(isPendingPublicLink(link)).toBe(true);
  });

  it("does not mark confirmed links or other link types as pending", () => {
    const confirmed: ClusterLink = {
      name: "target",
      type: "public",
      config: { "volatile.addresses": "10.0.0.2:8443" },
    };
    const bidirectional: ClusterLink = {
      name: "target",
      type: "bidirectional",
    };

    expect(isPendingPublicLink(confirmed)).toBe(false);
    expect(isPendingPublicLink(bidirectional)).toBe(false);
  });
});

describe("getCertificateFingerprint", () => {
  it("hashes the decoded certificate bytes as lowercase SHA-256 hex", async () => {
    const digest = vi
      .fn()
      .mockResolvedValue(Uint8Array.from([0, 15, 255]).buffer);
    vi.stubGlobal("crypto", { subtle: { digest } });

    const fingerprint = await getCertificateFingerprint(
      "-----BEGIN CERTIFICATE-----\nYWJj\n-----END CERTIFICATE-----",
    );

    expect(digest).toHaveBeenCalledWith(
      "SHA-256",
      Uint8Array.from([97, 98, 99]),
    );
    expect(fingerprint).toBe("000fff");
    vi.unstubAllGlobals();
  });
});

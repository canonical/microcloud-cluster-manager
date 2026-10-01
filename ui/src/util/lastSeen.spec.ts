import { getLastSeenLabel } from "./lastSeen";

describe("getLastSeenLabel", () => {
  it("shows minutes for the first few minutes", () => {
    expect(getLastSeenLabel(0)).toBe("seen 1 minute ago");
    expect(getLastSeenLabel(1)).toBe("seen 1 minute ago");
    expect(getLastSeenLabel(4)).toBe("seen 4 minutes ago");
  });

  it("shows the last hour between 5 and 59 minutes", () => {
    expect(getLastSeenLabel(5)).toBe("seen in last hour");
    expect(getLastSeenLabel(59)).toBe("seen in last hour");
  });

  it("shows hours after the first hour", () => {
    expect(getLastSeenLabel(60)).toBe("seen 1 hour ago");
    expect(getLastSeenLabel(119)).toBe("seen 1 hour ago");
    expect(getLastSeenLabel(180)).toBe("seen 3 hours ago");
    expect(getLastSeenLabel(3 * 24 * 60)).toBe("seen 72 hours ago");
  });
});

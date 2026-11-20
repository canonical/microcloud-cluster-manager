import { pluralize } from "./helpers";

describe("pluralize", () => {
  it.each([
    [0, "replicators"],
    [1, "replicator"],
    [2, "replicators"],
  ])("pluralizes regular nouns for count %i", (count, expected) => {
    expect(pluralize("replicator", count)).toBe(expected);
  });

  it.each([
    [0, "image registries"],
    [1, "image registry"],
    [2, "image registries"],
  ])("uses an explicit plural for count %i", (count, expected) => {
    expect(pluralize("image registry", count, "image registries")).toBe(
      expected,
    );
  });
});

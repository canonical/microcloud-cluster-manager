import { thresholdSchema } from "./thresholdValidation";

describe("thresholdSchema", () => {
  it("accepts whole percentages between 1 and 100", () => {
    expect(thresholdSchema.isValidSync(1)).toBe(true);
    expect(thresholdSchema.isValidSync(80)).toBe(true);
    expect(thresholdSchema.isValidSync(100)).toBe(true);
  });

  it("accepts an unset value", () => {
    expect(thresholdSchema.isValidSync(undefined)).toBe(true);
  });

  it("rejects values the API ignores or can't store", () => {
    expect(thresholdSchema.isValidSync(0)).toBe(false);
    expect(thresholdSchema.isValidSync(-5)).toBe(false);
    expect(thresholdSchema.isValidSync(101)).toBe(false);
    expect(thresholdSchema.isValidSync(85.5)).toBe(false);
  });
});

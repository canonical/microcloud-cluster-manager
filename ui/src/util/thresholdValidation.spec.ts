import { thresholdSchema, toThresholdFormValue } from "./thresholdValidation";

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

describe("toThresholdFormValue", () => {
  it("leaves a threshold that was never set empty", () => {
    expect(toThresholdFormValue(0, 80)).toBeUndefined();
  });

  it("keeps a stored threshold", () => {
    expect(toThresholdFormValue(42, 80)).toBe(42);
  });

  it("uses the fallback while the cluster is loading", () => {
    expect(toThresholdFormValue(undefined, 80)).toBe(80);
  });
});

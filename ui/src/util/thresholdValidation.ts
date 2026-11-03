import * as Yup from "yup";

const THRESHOLD_ERROR = "Enter a whole number between 1 and 100";

// The API stores thresholds as whole percentages and ignores a value of 0, so
// only accept integers between 1 and 100.
export const thresholdSchema = Yup.number()
  .integer(THRESHOLD_ERROR)
  .min(1, THRESHOLD_ERROR)
  .max(100, THRESHOLD_ERROR);

// The API reports 0 for a threshold that was never set. Show that as an empty
// field, which the schema accepts and the form leaves out of the request, so
// editing another setting doesn't fail validation or store a threshold.
export const toThresholdFormValue = (threshold?: number): number | undefined =>
  threshold === 0 ? undefined : threshold;

import type { Cluster, StatusDistribution } from "types/cluster";
import type { LxdIdentity } from "types/permissions";

interface ErrorResponse {
  error_code: number;
  error: string;
}

export class FetchError extends Error {
  response: ErrorResponse;

  constructor(response: ErrorResponse) {
    super(response.error);
    this.response = response;
  }
}

export const handleResponse = async (response: Response) => {
  if (!response.ok) {
    try {
      const result = (await response.json()) as ErrorResponse;
      throw new FetchError(result);
    } catch (err) {
      // If the error response body is empty or not JSON, create a generic error response
      if (err instanceof FetchError) {
        throw err;
      }
      console.error(
        `Failed to parse error response from ${response.url}: ${err instanceof Error ? err.message : String(err)}`,
      );
      const errorResponse: ErrorResponse = {
        error_code: response.status,
        error: `HTTP ${response.status}: ${response.statusText}`,
      };
      throw new FetchError(errorResponse);
    }
  }
  try {
    return (await response.json()) as unknown;
  } catch (err) {
    // If the success response body is empty or not JSON, throw a descriptive error
    console.error(
      `Failed to parse success response from ${response.url}: ${err instanceof Error ? err.message : String(err)}`,
    );
    const errorResponse: ErrorResponse = {
      error_code: response.status,
      error: `Invalid JSON response: ${err instanceof Error ? err.message : "Unknown error"}`,
    };
    throw new FetchError(errorResponse);
  }
};

export const isDimensionBelow = (
  dimension: number,
  mode: "width" | "height" = "width",
): boolean =>
  mode === "width"
    ? window.innerWidth < dimension
    : window.innerHeight < dimension;

export function getMinutesSinceLastHeartbeat(cluster: Cluster): number {
  const lastSeenTime = Date.parse(cluster.last_status_update_at);
  return Math.floor((Date.now() - lastSeenTime) / 60000);
}

export function getSecondsSinceLastHeartbeat(cluster?: Cluster): number {
  if (!cluster) {
    return 0;
  }
  const lastSeenTime = Date.parse(cluster.last_status_update_at);
  return Math.floor((Date.now() - lastSeenTime) / 1000);
}

export const humanFileSize = (bytes: number): string => {
  if (Math.abs(bytes) < 1000) {
    return `${bytes} B`;
  }

  const units = ["KiB", "MiB", "GiB", "TiB", "PiB", "EiB", "ZiB", "YiB"];
  let u = -1;

  do {
    bytes /= 1024;
    ++u;
  } while (
    Math.round(Math.abs(bytes) * 10) / 10 >= 1000 &&
    u < units.length - 1
  );

  return `${bytes.toFixed(1)} ${units[u]}`;
};

export const isoTimeToString = (isoTime: string): string => {
  const date = new Date(isoTime);
  if (date.getTime() === 0) {
    return "Never";
  }

  return date.toLocaleDateString(undefined, {
    year: "numeric",
    month: "short",
    day: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
};

export const logout = (): void => {
  void fetch(`/oidc/logout`).then(() => {
    if (!window.location.href.includes("/ui/login")) {
      window.location.href = "/ui/login";
    }
  });
};

export const convertToISOFormat = (datetimeLocalString: string) => {
  // Split the datetime-local string into date and time parts
  const [datePart, timePart] = datetimeLocalString.split("T");

  // Split time part into hours and minutes
  const [hours, minutes] = timePart.split(":");

  // Create a new Date object with the parts
  const date = new Date(`${datePart}T${hours}:${minutes}:00`);

  // Return the ISO 8601 formatted string
  return date.toISOString();
};

// this works only for words that form the plural by adding an "s" at the end
export const pluralize = (word: string, count: number): string => {
  return count === 1 ? word : `${word}s`;
};

export const statusCount = (
  distribution: StatusDistribution[],
  status: string,
) => {
  return distribution.find((item) => item.status === status)?.count ?? 0;
};

export const handleSettledResult = (
  results: PromiseSettledResult<unknown>[],
): void => {
  const error = results.find((res) => res.status === "rejected")?.reason as
    | Error
    | undefined;

  if (error) {
    throw error;
  }
};

export const capitalizeFirstLetter = (val: string): string =>
  val.charAt(0).toUpperCase() + val.slice(1);

type EtagObject = LxdIdentity;

interface SortableObject {
  access_entitlements?: string[];
  locations?: string[];
  used_by?: string[];
}

export const ensureStableSorting = (data: SortableObject) => {
  // these data fields can be randomly sorted in the API response,
  // we sort them here to have consistent results
  // and avoid forms from re-initializing based on different order
  if (data.access_entitlements) {
    data.access_entitlements.sort();
  }
  if (data.locations) {
    data.locations.sort();
  }
  if (data.used_by) {
    data.used_by.sort();
  }
  return data as EtagObject;
};

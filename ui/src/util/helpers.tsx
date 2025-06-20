import { Cluster, StatusDistribution } from "types/cluster";

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
    const result = (await response.json()) as ErrorResponse;
    throw new FetchError(result);
  }
  return response.json() as unknown;
};

export const isWidthBelow = (width: number): boolean =>
  window.innerWidth < width;

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
  const postLogoutPath = "/ui/login";

  void fetch(`/oidc/logout?next=${postLogoutPath}`).then(() => {
    if (!window.location.href.includes("/ui/login")) {
      window.location.href = "/ui/login";
    }
  });
};

export const getElementAbsoluteHeight = (element: HTMLElement) => {
  if (!element) {
    return 0;
  }
  const style = window.getComputedStyle(element);
  const margin = parseFloat(style.marginTop) + parseFloat(style.marginBottom);
  const padding =
    parseFloat(style.paddingTop) + parseFloat(style.paddingBottom);
  return element.offsetHeight + margin + padding + 1;
};

export const getAbsoluteHeightBelowById = (belowId: string): number => {
  const element = belowId ? document.getElementById(belowId) : undefined;
  if (!element) {
    return 0;
  }
  return getElementAbsoluteHeight(element);
};

export const getParentsBottomSpacing = (element: Element): number => {
  let sum = 0;
  while (element.parentElement) {
    element = element.parentElement;
    const style = window.getComputedStyle(element);
    const margin = parseInt(style.marginBottom);
    const padding = parseInt(style.paddingBottom);
    sum += margin + padding;
  }
  return sum;
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

import { Cluster } from "types/cluster";

interface ErrorResponse {
  error_code: number;
  error: string;
}

export const handleResponse = async (response: Response) => {
  if (!response.ok) {
    // eslint-disable-next-line @typescript-eslint/no-unsafe-assignment
    const result: ErrorResponse = await response.json();
    throw Error(result.error);
  }
  return response.json();
};

export const isWidthBelow = (width: number): boolean =>
  window.innerWidth < width;

export function getMinutesSinceLastHeartbeat(cluster: Cluster): number {
  const lastSeenTime = Date.parse(cluster.last_status_update_at);
  return Math.floor((Date.now() - lastSeenTime) / 60000);
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

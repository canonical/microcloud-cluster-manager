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

export function getHeartbeatStatus(cluster: Cluster): string {
  const lastSeenTime = Date.parse(cluster.last_status_update_at);
  const lastHeartbeat = Math.floor((Date.now() - lastSeenTime) / 60000); //Value in Minutes

  if (lastHeartbeat <= 1) {
    return `< 1 minute ago`;
  } else if (lastHeartbeat < 5) {
    return `${lastHeartbeat} minutes ago`;
  }

  return "Unresponsive";
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

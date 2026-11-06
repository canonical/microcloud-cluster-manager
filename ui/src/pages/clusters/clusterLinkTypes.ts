import type { ClusterLink } from "types/cluster";

export const LINK_TYPES: {
  value: ClusterLink["type"];
  label: string;
  help: string;
}[] = [
  {
    value: "bidirectional",
    label: "Bidirectional",
    help: "Connect two LXD clusters with access in both directions. Required for replication.",
  },
  {
    value: "unidirectional",
    label: "Unidirectional",
    help: "Access another LXD cluster without granting access in return. Often used to fetch images.",
  },
  {
    value: "public",
    label: "Public",
    help: "Access another LXD cluster without authentication or return access. Used for public image fetching.",
  },
];

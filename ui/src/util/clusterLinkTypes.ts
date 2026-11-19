import type { ResourceIconType } from "components/ResourceIcon";
import type { ClusterLink } from "types/cluster";

export const LINK_TYPES: {
  value: ClusterLink["type"];
  label: string;
  help: string;
  tooltip: string;
}[] = [
  {
    value: "bidirectional",
    label: "Bidirectional",
    help: "Connect two LXD clusters with access in both directions. Required for replication.",
    tooltip:
      "Bidirectional links connect two LXD clusters with access in both directions. Required for replication.",
  },
  {
    value: "unidirectional",
    label: "Unidirectional",
    help: "Access another LXD cluster without granting access in return. Often used to fetch images.",
    tooltip:
      "Unidirectional links access another LXD cluster without granting access in return. Often used to fetch images.",
  },
  {
    value: "public",
    label: "Public",
    help: "Access another LXD cluster without authentication or return access. Used for public image fetching.",
    tooltip:
      "Access another LXD cluster without authentication or return access. Used for public image fetching and anonymous read access.",
  },
];

export const getLinkIconType = (
  type: ClusterLink["type"],
): ResourceIconType => {
  switch (type) {
    case "public":
      return "cluster-link-public";
    case "unidirectional":
      return "cluster-link-unidirectional";
    default:
      return "cluster-link-bidirectional";
  }
};

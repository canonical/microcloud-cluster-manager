import type { IconName } from "@canonical/ds-assets";
import type { FC } from "react";
import DsIcon from "components/DsIcon";

export type ResourceIconType =
  | "cluster"
  | "cluster-link-bidirectional"
  | "cluster-link-unidirectional"
  | "cluster-link-public";

// Icons available in @canonical/ds-assets — rendered via DsIcon (<svg><use>).
// Move entries here from vanillaResourceIcons as ds-assets gains coverage.
const dsResourceIcons: Record<ResourceIconType, IconName> = {
  cluster: "cluster-host",
  "cluster-link-bidirectional": "swap",
  "cluster-link-unidirectional": "arrow-right",
  "cluster-link-public": "exposed",
};

interface Props {
  type: ResourceIconType;
  className?: string;
}

const ResourceIcon: FC<Props> = ({ type, className }) => {
  return <DsIcon icon={dsResourceIcons[type]} className={className} />;
};

export default ResourceIcon;

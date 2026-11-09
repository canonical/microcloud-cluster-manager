import type { FC } from "react";
import DsIcon from "components/DsIcon";

export type ResourceIconType =
  | "cluster"
  | "cluster-link-bidirectional"
  | "cluster-link-unidirectional"
  | "cluster-link-public"
  | "cluster-link"
  | "setting";

type DsIconName =
  | "cluster-host"
  | "swap"
  | "arrow-right"
  | "exposed"
  | "applications"
  | "settings";

const dsResourceIcons: Record<ResourceIconType, DsIconName> = {
  cluster: "cluster-host",
  "cluster-link-bidirectional": "swap",
  "cluster-link-unidirectional": "arrow-right",
  "cluster-link-public": "exposed",
  "cluster-link": "applications",
  setting: "settings",
};

interface Props {
  type: ResourceIconType;
  className?: string;
}

const ResourceIcon: FC<Props> = ({ type, className }) => {
  return <DsIcon icon={dsResourceIcons[type]} className={className} />;
};

export default ResourceIcon;

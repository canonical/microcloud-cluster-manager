import { Icon } from "@canonical/react-components";
import type { FC } from "react";

export type ResourceIconType = "cluster" | "cluster-link" | "setting";

const resourceIcons: Record<ResourceIconType, string> = {
  cluster: "cluster-host",
  "cluster-link": "applications",
  setting: "settings",
};

interface Props {
  type: ResourceIconType;
  className?: string;
}

const ResourceIcon: FC<Props> = ({ type, className }) => {
  return <Icon name={resourceIcons[type]} className={className} />;
};

export default ResourceIcon;

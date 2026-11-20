import type { FC } from "react";
import { pluralize } from "util/helpers";

interface Props {
  usedBy: string[];
}

const ClusterLinkUsedBy: FC<Props> = ({ usedBy }) => {
  const replicatorCount = usedBy.filter((path) =>
    path.startsWith("/1.0/replicators/"),
  ).length;
  const imageRegistryCount = usedBy.filter((path) =>
    path.startsWith("/1.0/image-registries/"),
  ).length;
  const projectCount = usedBy.filter((path) =>
    path.startsWith("/1.0/projects/"),
  ).length;

  return (
    <>
      {replicatorCount > 0 && (
        <div>
          {`${replicatorCount} ${pluralize("replicator", replicatorCount)}`}
        </div>
      )}
      {imageRegistryCount > 0 && (
        <div>
          {`${imageRegistryCount} ${pluralize("image registry", imageRegistryCount, "image registries")}`}
        </div>
      )}
      {projectCount > 0 && (
        <div>{`${projectCount} ${pluralize("project", projectCount)}`}</div>
      )}
    </>
  );
};

export default ClusterLinkUsedBy;

import type { FC } from "react";
import type { Cluster } from "types/cluster";
import { Icon, usePortal } from "@canonical/react-components";
import { Link } from "react-router-dom";
import { ClusterDiskModal } from "pages/clusters/metrics/ClusterDiskModal";

interface Props {
  cluster: Cluster;
}

export const ClusterDiskInfo: FC<Props> = ({ cluster }: Props) => {
  const { openPortal, closePortal, isOpen, Portal } = usePortal();

  return (
    <>
      {isOpen && (
        <Portal>
          <ClusterDiskModal cluster={cluster} onClose={closePortal} />
        </Portal>
      )}
      <Link onClick={openPortal} to="#">
        <Icon name="information" />
      </Link>
    </>
  );
};

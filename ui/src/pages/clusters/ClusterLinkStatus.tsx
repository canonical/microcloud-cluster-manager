import type { FC } from "react";
import { Icon, Spinner } from "@canonical/react-components";
import type { ClusterLink, StatusCaption } from "types/cluster";
import { useClusterLinkState } from "context/useClusters";
import { useIdentities } from "context/useIdentities";
import { getClusterLinksStatus, getLinkIdentity } from "util/clusterLinks";

interface Props {
  cluster: string;
  link: ClusterLink;
}

const STATUS_ICONS: Record<StatusCaption, string> = {
  Reachable: "status-succeeded-small",
  Unreachable: "status-failed-small",
  Pending: "status-queued-small",
};

const ClusterLinkStatus: FC<Props> = ({ cluster, link }) => {
  const { data: state, isLoading } = useClusterLinkState(cluster, link.name);
  const { data: identities = [] } = useIdentities(cluster);

  if (isLoading) {
    return <Spinner className="status-spinner" />;
  }

  const identity = getLinkIdentity(identities, link.name);
  const status = getClusterLinksStatus(identity, state, link);

  return (
    <>
      <Icon name={STATUS_ICONS[status]} className="status-icon" />
      {status}
    </>
  );
};

export default ClusterLinkStatus;

import type { FC } from "react";
import { Spinner } from "@canonical/react-components";
import StatusIcon from "components/StatusIcon";
import { useClusterIdentities } from "context/useClusterIdentities";
import { useClusterLinkState } from "context/useClusterLinks";
import type { ClusterLink, StatusCaption } from "types/cluster";
import { getClusterLinksStatus, getLinkIdentity } from "util/clusterLinks";
import type { StatusTone } from "util/statusColors";

interface Props {
  cluster: string;
  link: ClusterLink;
}

const STATUS_ICONS: Record<StatusCaption, StatusTone> = {
  Reachable: "positive",
  Unreachable: "negative",
  Pending: "neutral",
};

const ClusterLinkStatus: FC<Props> = ({ cluster, link }) => {
  const { data: state, isLoading } = useClusterLinkState(cluster, link.name);
  const { data: identities = [] } = useClusterIdentities(cluster);

  if (isLoading) {
    return <Spinner className="status-spinner" />;
  }

  const identity = getLinkIdentity(identities, link.name);
  const status = getClusterLinksStatus(identity, state, link);

  return (
    <>
      <StatusIcon tone={STATUS_ICONS[status]} className="status-icon" />
      {status}
    </>
  );
};

export default ClusterLinkStatus;

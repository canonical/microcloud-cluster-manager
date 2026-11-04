import { type FC } from "react";
import { Spinner } from "@canonical/react-components";
import type { TooltipRow } from "components/RichTooltipRow";
import { RichTooltipTable } from "components/RichTooltipTable";
import ItemName from "components/ItemName";
import { useClusterLink } from "context/useClusters";
import ClusterLinkStatus from "./ClusterLinkStatus";
import ResourceLabel from "components/ResourceLabel";
import ClusterLinkAddresses from "./ClusterLinkAddresses";
import { useIdentities } from "context/useIdentities";
import type { ClusterLink } from "types/cluster";
import { getLinkIdentity } from "util/clusterLinks";

interface Props {
  cluster: string;
  clusterLink: ClusterLink;
}

const ClusterLinkRichTooltip: FC<Props> = ({ cluster, clusterLink }) => {
  const clusterLinkName = clusterLink.name;
  const { data: link, isLoading: isLinkLoading } = useClusterLink(
    cluster,
    clusterLinkName,
  );
  const { data: identities = [] } = useIdentities(cluster);

  if (!link && !isLinkLoading) {
    return (
      <>
        Cluster link{" "}
        <ResourceLabel
          type={`cluster-link-${clusterLink.type}`}
          value={clusterLinkName}
        />{" "}
        not found
      </>
    );
  }

  const identity = getLinkIdentity(identities, clusterLinkName);
  const authGroupCount = identity?.groups?.length ?? 0;

  const rows: TooltipRow[] = [
    {
      title: "Cluster link",
      value: link ? <ItemName item={{ name: clusterLinkName }} /> : <Spinner />,
      valueTitle: clusterLinkName,
    },
    {
      title: "Status",
      value: link ? <ClusterLinkStatus cluster={cluster} link={link} /> : "-",
      className: "status-row",
    },
    {
      title: "Description",
      value: link?.description || "-",
      valueTitle: link?.description || "",
    },
    {
      title: "Type",
      value: link?.type || "-",
    },
    {
      title: "Addresses",
      value: link ? <ClusterLinkAddresses clusterLink={link} /> : "-",
    },
    {
      title: "Auth groups",
      value: authGroupCount,
    },
  ];

  return (
    <RichTooltipTable rows={rows} className="cluster-link-rich-tooltip-table" />
  );
};

export default ClusterLinkRichTooltip;

import { type FC } from "react";
import type { UseQueryResult } from "@tanstack/react-query";
import DocLink from "components/DocLink";
import PanelSection from "components/PanelSection";
import ClusterAuthGroupSelector from "pages/clusters/ClusterAuthGroupSelector";
import type { LxdAuthGroup, LxdIdentity } from "types/permissions";

export interface ClusterAuthGroupSelection {
  clusterName: string;
  query: UseQueryResult<LxdAuthGroup[]>;
  selectedGroups: string[];
  onChange: (groups: string[]) => void;
  identityQuery?: UseQueryResult<LxdIdentity>;
  unavailableReason?: string;
}

type Props = {
  target: ClusterAuthGroupSelection;
  disabled: boolean;
} & (
  | { isBidirectional: true; source: ClusterAuthGroupSelection }
  | { isBidirectional: false; source?: never }
);

const ClusterLinkAuthGroupsSection: FC<Props> = ({
  target,
  source,
  isBidirectional,
  disabled,
}) => (
  <PanelSection header="Auth groups">
    <p className="u-text--muted">
      {isBidirectional
        ? "Control the permissions that the source and target have to act on each other."
        : "Control the permissions that the source has to act on the target."}{" "}
      <br />
      <DocLink docPath="/howto/cluster_links_manage/#manage-cluster-link-permissions">
        Learn more about cluster link permissions in the docs.
      </DocLink>
    </p>

    {isBidirectional && (
      <ClusterAuthGroupSelector
        key={source.clusterName}
        {...source}
        direction="Source"
        disabled={disabled}
      />
    )}
    <ClusterAuthGroupSelector
      key={target.clusterName}
      {...target}
      direction="Target"
      disabled={disabled}
    />
  </PanelSection>
);

export default ClusterLinkAuthGroupsSection;

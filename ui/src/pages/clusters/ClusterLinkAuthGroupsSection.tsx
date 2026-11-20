import { useId, type FC } from "react";
import { Button, Notification, Spinner } from "@canonical/react-components";
import type { UseQueryResult } from "@tanstack/react-query";
import DocLink from "components/DocLink";
import PanelSection from "components/PanelSection";
import GroupSelection from "pages/clusters/GroupSelection";
import type { LxdAuthGroup, LxdIdentity } from "types/permissions";

interface Selection {
  clusterName: string;
  query: UseQueryResult<LxdAuthGroup[]>;
  selectedGroups: string[];
  onChange: (groups: string[]) => void;
  identityQuery?: UseQueryResult<LxdIdentity>;
  unavailableReason?: string;
}

type Props = {
  source: Selection;
  disabled: boolean;
} & (
  | { isBidirectional: true; target: Selection }
  | { isBidirectional: false; target?: never }
);

interface SelectionProps extends Selection {
  direction: "Source" | "Target";
  disabled: boolean;
}

const AuthGroupSelection: FC<SelectionProps> = ({
  direction,
  clusterName,
  query,
  selectedGroups,
  onChange,
  identityQuery,
  unavailableReason,
  disabled,
}) => {
  const headingId = useId();
  const searchLabelId = useId();
  const selected = new Set(selectedGroups);
  const initial = new Set(identityQuery?.data?.groups ?? []);
  const modified = new Set(
    [...selected, ...initial].filter(
      (group) => selected.has(group) !== initial.has(group),
    ),
  );

  return (
    <div role="group" aria-labelledby={headingId}>
      <p id={headingId} className="u-no-margin--bottom">
        {direction}
      </p>
      <span id={searchLabelId} className="u-off-screen">
        Search {direction.toLowerCase()} auth groups
      </span>
      <p className="u-text--muted">
        {direction === "Source"
          ? "Control how the source can act on the target."
          : "Control how the target can act on the source."}
      </p>
      {unavailableReason ? (
        <Notification severity="caution" title="Auth groups unavailable">
          {unavailableReason}
        </Notification>
      ) : !clusterName ? (
        <p>Select a target cluster to load its auth groups.</p>
      ) : identityQuery?.isError ? (
        <Notification
          severity="negative"
          title={`Loading ${direction.toLowerCase()} identity failed`}
          actions={[
            <Button
              key="retry"
              onClick={() => void identityQuery.refetch()}
              disabled={disabled || identityQuery.isFetching}
            >
              Retry
            </Button>,
          ]}
        >
          {identityQuery.error.message}
        </Notification>
      ) : query.isError ? (
        <Notification
          severity="negative"
          title={`Loading ${direction.toLowerCase()} auth groups failed`}
          actions={[
            <Button
              key="retry"
              onClick={() => void query.refetch()}
              disabled={disabled || query.isFetching}
            >
              Retry
            </Button>,
          ]}
        >
          {query.error.message}
        </Notification>
      ) : identityQuery?.isPending || query.isPending ? (
        <Spinner text={`Loading ${direction.toLowerCase()} auth groups...`} />
      ) : (
        <GroupSelection
          groups={query.data}
          modifiedGroups={modified}
          parentItemName="cluster link"
          selectedGroups={selected}
          preselectedGroups={identityQuery ? initial : undefined}
          setSelectedGroups={onChange}
          toggleGroup={(group) => {
            onChange(
              selected.has(group)
                ? selectedGroups.filter((name) => name !== group)
                : [...selectedGroups, group],
            );
          }}
          scrollDependencies={[selectedGroups]}
          hasScrollableTable={false}
          disabled={disabled}
          labelledBy={headingId}
          searchLabelledBy={searchLabelId}
        />
      )}
    </div>
  );
};

const ClusterLinkAuthGroupsSection: FC<Props> = (props) => (
  <PanelSection header="Auth groups">
    <p className="u-text--muted">
      {props.isBidirectional
        ? "Control the permissions that the source and target have to act on each other."
        : "Control the permissions that the source has to act on the target."}{" "}
      <br />
      <DocLink docPath="/howto/cluster_links_manage/#manage-cluster-link-permissions">
        Learn more about cluster link permissions in the docs.
      </DocLink>
    </p>
    <AuthGroupSelection
      key={props.source.clusterName}
      {...props.source}
      direction="Target"
      disabled={props.disabled}
    />
    {props.isBidirectional && (
      <AuthGroupSelection
        {...props.target}
        direction="Source"
        disabled={props.disabled}
      />
    )}
  </PanelSection>
);

export default ClusterLinkAuthGroupsSection;

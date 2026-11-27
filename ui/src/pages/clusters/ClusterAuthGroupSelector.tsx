import { useId, type FC } from "react";
import { Notification, Button, Spinner } from "@canonical/react-components";
import type { ClusterAuthGroupSelection } from "pages/clusters/ClusterLinkAuthGroupsSection";
import GroupSelection from "pages/clusters/GroupSelection";

interface Props extends ClusterAuthGroupSelection {
  direction: "Source" | "Target";
  disabled: boolean;
}

const ClusterAuthGroupSelector: FC<Props> = ({
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
        {direction === "Target"
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

export default ClusterAuthGroupSelector;

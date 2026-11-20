import {
  CheckboxInput,
  ConfirmationButton,
  useToastNotification,
} from "@canonical/react-components";
import { useQueryClient } from "@tanstack/react-query";
import { deleteClusterLink, deleteClusterLinks } from "api/clusters";
import type { FC } from "react";
import { useState } from "react";
import DsIcon from "components/DsIcon";
import { invalidateClusterLinks } from "util/clusterLinkQueries";
import type { ClusterConnection } from "util/clusterLinks";
import {
  canRemoveTargetEnd,
  hasTargetEnd,
  isBidirectional,
} from "util/clusterLinks";
import { pluralize } from "util/helpers";

interface Props {
  cluster: string;
  links: ClusterConnection[];
  hasLabel?: boolean;
  disabled?: boolean;
  onStart?: () => void;
  onFinish?: () => void;
}

const DeleteClusterLinksButton: FC<Props> = ({
  cluster,
  links,
  hasLabel = true,
  disabled = false,
  onStart,
  onFinish,
}) => {
  const queryClient = useQueryClient();
  const toastNotification = useToastNotification();
  const [isLoading, setLoading] = useState(false);
  const [deleteBothEnds, setDeleteBothEnds] = useState(true);

  const isSingle = links.length === 1;
  const linksWithTargetEnd = links.filter(({ link }) => hasTargetEnd(link));
  const cleanableLinks = links.filter(({ link, target }) =>
    canRemoveTargetEnd(link, target),
  );
  const canCleanTargets = cleanableLinks.length > 0;
  const uncleanableCount = linksWithTargetEnd.length - cleanableLinks.length;
  const removesTargetEnds = deleteBothEnds && canCleanTargets;
  const checkboxLabel = linksWithTargetEnd.some(({ link }) =>
    isBidirectional(link),
  )
    ? "Delete links on both source and target cluster"
    : "Also delete identity on the target cluster";
  const title = isSingle ? "Delete cluster link" : "Delete cluster links";
  const linkCount = isSingle ? (
    <>
      cluster link <strong>{links[0].link.name}</strong>
    </>
  ) : (
    <strong>
      {links.length} {pluralize("cluster link", links.length)}
    </strong>
  );

  const getSuccessMessage = () => {
    if (!isSingle) {
      return <>Deleted {linkCount}.</>;
    }
    const { link, target } = links[0];
    if (removesTargetEnds && target) {
      return (
        <>
          Deleted cluster link between <strong>{cluster}</strong> and{" "}
          <strong>{target.name}</strong>.
        </>
      );
    }
    return (
      <>
        Deleted cluster link <strong>{link.name}</strong>.
      </>
    );
  };

  const handleDelete = async () => {
    setLoading(true);
    onStart?.();
    const requests = links.map(({ link, target }) => ({
      name: link.name,
      removeReverseEdge: removesTargetEnds && canRemoveTargetEnd(link, target),
    }));
    try {
      if (isSingle) {
        await deleteClusterLink(
          cluster,
          requests[0].name,
          requests[0].removeReverseEdge,
        );
      } else {
        await deleteClusterLinks(cluster, requests);
      }
      toastNotification.success(getSuccessMessage());
    } catch (error) {
      toastNotification.failure(
        isSingle
          ? `Unable to delete cluster link ${links[0].link.name}.`
          : "Cluster link deletion failed.",
        error,
      );
    }

    await invalidateClusterLinks(queryClient);
    setDeleteBothEnds(true);
    setLoading(false);
    onFinish?.();
  };

  const getUncleanableNote = () => {
    if (uncleanableCount === 0) {
      return null;
    }
    if (isSingle) {
      return "The target cluster is not managed by cluster manager or has no registered tunnel. Only this side of the link is removed; the target cluster must be cleaned up separately.";
    }
    return `${uncleanableCount} of the selected ${pluralize("link", linksWithTargetEnd.length)} ${uncleanableCount === 1 ? "has a target" : "have targets"} that are not managed by cluster manager or have no registered tunnel. Those links are only removed from this side.`;
  };

  return (
    <ConfirmationButton
      className="u-no-margin--bottom has-icon"
      appearance={hasLabel ? "default" : "base"}
      loading={isLoading}
      disabled={disabled || undefined}
      aria-label={
        hasLabel
          ? undefined
          : isSingle
            ? `Delete cluster link ${links[0].link.name}`
            : title
      }
      onHoverText={hasLabel ? undefined : "Delete"}
      confirmationModalProps={{
        title,
        children: (
          <>
            <p>
              This will permanently delete {linkCount}. This action cannot be
              undone.
            </p>
            {linksWithTargetEnd.length > 0 && (
              <CheckboxInput
                id="delete-cluster-links-both-ends"
                label={checkboxLabel}
                checked={removesTargetEnds}
                disabled={!canCleanTargets}
                onChange={() => {
                  setDeleteBothEnds((value) => !value);
                }}
              />
            )}
            {getUncleanableNote() && <p>{getUncleanableNote()}</p>}
          </>
        ),
        confirmButtonLabel: "Delete",
        close: () => {
          setDeleteBothEnds(true);
        },
        onConfirm: () => void handleDelete(),
      }}
      shiftClickEnabled
    >
      <DsIcon icon="delete" />
      {hasLabel && (
        <span>Delete {pluralize("cluster link", links.length)}</span>
      )}
    </ConfirmationButton>
  );
};

export default DeleteClusterLinksButton;

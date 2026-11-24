import {
  CheckboxInput,
  ConfirmationButton,
  useToastNotification,
} from "@canonical/react-components";
import { useQueryClient } from "@tanstack/react-query";
import { deleteClusterLink } from "api/clusters";
import type { FC } from "react";
import { useState } from "react";
import DsIcon from "components/DsIcon";
import type { ClusterConnection } from "util/clusterLinks";
import {
  canRemoveTargetEnd,
  hasTargetEnd,
  invalidateClusterLinksAndIdentities,
  isBidirectional,
} from "util/clusterLinks";

interface Props {
  entry: ClusterConnection;
}

const DeleteClusterLinkBtn: FC<Props> = ({ entry }) => {
  const { link, source, target } = entry;
  const queryClient = useQueryClient();
  const toastNotification = useToastNotification();
  const [isLoading, setLoading] = useState(false);
  const [deleteBothEnds, setDeleteBothEnds] = useState(true);
  const canCleanTarget = canRemoveTargetEnd(link, target);
  const targetToClean = deleteBothEnds && canCleanTarget ? target : undefined;

  const handleDelete = async (): Promise<void> => {
    setLoading(true);
    try {
      await deleteClusterLink(source.name, link.name, !!targetToClean);
      toastNotification.success(
        targetToClean && isBidirectional(link) ? (
          <>
            Deleted cluster link between <strong>{source.name}</strong> and{" "}
            <strong>{targetToClean.name}</strong>.
          </>
        ) : (
          <>
            Deleted cluster link <strong>{link.name}</strong>.
          </>
        ),
      );
    } catch (error) {
      toastNotification.failure(
        `Unable to delete cluster link ${link.name}.`,
        error,
      );
    } finally {
      await invalidateClusterLinksAndIdentities(
        queryClient,
        source.name,
        link,
        targetToClean?.name,
      );
      setDeleteBothEnds(true);
      setLoading(false);
    }
  };

  return (
    <ConfirmationButton
      appearance="base"
      className="u-no-margin--bottom has-icon"
      loading={isLoading}
      aria-label={`Delete cluster link ${link.name}`}
      onHoverText="Delete"
      confirmationModalProps={{
        title: "Delete cluster link",
        children: (
          <>
            <p>
              This will permanently delete cluster link{" "}
              <strong>{link.name}</strong>. This action cannot be undone.
            </p>
            {hasTargetEnd(link) &&
              (canCleanTarget ? (
                <CheckboxInput
                  id="delete-cluster-link-both-ends"
                  label={
                    isBidirectional(link)
                      ? "Delete links on both source and target cluster"
                      : "Also delete identity on the target cluster"
                  }
                  checked={deleteBothEnds}
                  onChange={() => {
                    setDeleteBothEnds((value) => !value);
                  }}
                />
              ) : (
                <p>
                  Only this side of the link is removed. The linked cluster must
                  be cleaned up separately.
                </p>
              ))}
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
    </ConfirmationButton>
  );
};

export default DeleteClusterLinkBtn;

import {
  CheckboxInput,
  ConfirmationButton,
  useToastNotification,
} from "@canonical/react-components";
import { useQueryClient } from "@tanstack/react-query";
import { deleteClusterLink } from "api/clusters";
import type { FC } from "react";
import { useState } from "react";
import { queryKeys } from "util/queryKeys";
import type { Cluster, ClusterLink } from "types/cluster";
import DsIcon from "components/DsIcon";
import {
  canRemoveTargetEnd,
  hasTargetEnd,
  isBidirectional,
} from "util/clusterLinks";

interface Props {
  cluster: Cluster;
  link: ClusterLink;
  targetCluster?: Cluster;
  iconOnly?: boolean;
}

const DeleteClusterLinkBtn: FC<Props> = ({
  cluster,
  link,
  targetCluster,
  iconOnly = false,
}) => {
  const queryClient = useQueryClient();
  const toastNotification = useToastNotification();
  const [isLoading, setLoading] = useState(false);
  const [deleteBothEnds, setDeleteBothEnds] = useState(true);
  const canCleanTarget = canRemoveTargetEnd(link, targetCluster);

  const handleDelete = async () => {
    setLoading(true);
    const removeReverseEdge = deleteBothEnds && canCleanTarget;
    try {
      await deleteClusterLink(cluster.name, link.name, removeReverseEdge);

      toastNotification.success(
        removeReverseEdge ? (
          <>
            Deleted cluster link between <strong>{cluster.name}</strong> and{" "}
            <strong>{targetCluster?.name}</strong>.
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
    }

    await queryClient.invalidateQueries({
      queryKey: [queryKeys.clusters],
    });
    setDeleteBothEnds(true);
    setLoading(false);
  };

  return (
    <ConfirmationButton
      appearance={iconOnly ? "base" : undefined}
      className="u-no-margin--bottom has-icon"
      loading={isLoading}
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
                  id="delete-both-ends"
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
      {iconOnly ? (
        <span className="u-off-screen">Delete {link.name}</span>
      ) : (
        <span>Delete</span>
      )}
    </ConfirmationButton>
  );
};

export default DeleteClusterLinkBtn;

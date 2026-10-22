import {
  CheckboxInput,
  ConfirmationButton,
  Icon,
  useToastNotification,
} from "@canonical/react-components";
import { useQueryClient } from "@tanstack/react-query";
import { deleteClusterLink, fetchClusterLinks } from "api/clusters";
import type { FC } from "react";
import { useState } from "react";
import { queryKeys } from "util/queryKeys";
import type { Cluster, ClusterLink } from "types/cluster";
import { getReverseLink } from "util/clusterLinks";

interface Props {
  cluster: Cluster;
  link: ClusterLink;
  targetCluster?: Cluster;
}

const DeleteClusterLinkBtn: FC<Props> = ({ cluster, link, targetCluster }) => {
  const queryClient = useQueryClient();
  const toastNotification = useToastNotification();
  const [isLoading, setLoading] = useState(false);
  const [deleteBothEnds, setDeleteBothEnds] = useState(true);
  const isBidirectional = link.type.toLowerCase() === "bidirectional";
  const canCleanTarget = !!targetCluster?.tunnel_registered;

  const deleteTargetEnd = async (target: Cluster) => {
    const targetLinks = await fetchClusterLinks(target.name);
    const reverseLink = getReverseLink(targetLinks, cluster);
    if (!reverseLink) {
      throw new Error(
        `No cluster link pointing back at ${cluster.name} found on cluster ${target.name}.`,
      );
    }

    await deleteClusterLink(target.name, reverseLink.name);
  };

  const handleDelete = async () => {
    setLoading(true);
    const removeTargetEnd = isBidirectional && deleteBothEnds && canCleanTarget;
    try {
      await deleteClusterLink(cluster.name, link.name);

      if (targetCluster && removeTargetEnd) {
        await deleteTargetEnd(targetCluster);
      }

      toastNotification.success(
        removeTargetEnd ? (
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
            {isBidirectional &&
              (canCleanTarget ? (
                <CheckboxInput
                  id="delete-both-ends"
                  label="Delete bidirectional links on both source and target cluster"
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
      <Icon name="delete" />
      <span>Delete</span>
    </ConfirmationButton>
  );
};

export default DeleteClusterLinkBtn;

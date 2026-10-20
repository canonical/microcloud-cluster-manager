import {
  CheckboxInput,
  ConfirmationButton,
  Icon,
  useNotify,
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
  const notify = useNotify();
  const toastNotification = useToastNotification();
  const [isLoading, setLoading] = useState(false);
  const [deleteBothEnds, setDeleteBothEnds] = useState(false);

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
    const removeTargetEnd = deleteBothEnds && !!targetCluster;
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
      notify.failure(`Unable to delete cluster link ${link.name}.`, error);
    }

    await queryClient.invalidateQueries({
      queryKey: [queryKeys.clusters],
    });
    setDeleteBothEnds(false);
    setLoading(false);
  };

  return (
    <ConfirmationButton
      className="u-no-margin--bottom has-icon"
      loading={isLoading}
      confirmationModalProps={{
        title: "Confirm delete",
        children: (
          <>
            <p>
              Are you sure you want to delete the cluster link{" "}
              <strong>{link.name}</strong> on cluster{" "}
              <strong>{cluster.name}</strong>?
            </p>
            {targetCluster ? (
              <CheckboxInput
                id="delete-both-ends"
                label={
                  <>
                    Also delete the link on{" "}
                    <strong>{targetCluster.name}</strong>
                  </>
                }
                checked={deleteBothEnds}
                onChange={() => {
                  setDeleteBothEnds((value) => !value);
                }}
              />
            ) : (
              <p>
                Only this side of the link is removed. The linked cluster keeps
                its own cluster link and needs to be cleaned up separately.
              </p>
            )}
          </>
        ),
        confirmButtonLabel: "Confirm delete",
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

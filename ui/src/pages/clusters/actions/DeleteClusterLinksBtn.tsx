import {
  ConfirmationButton,
  Icon,
  useToastNotification,
} from "@canonical/react-components";
import { useQueryClient } from "@tanstack/react-query";
import type { FC } from "react";
import { useState } from "react";
import type { Cluster, ClusterLink } from "types/cluster";
import { getLinkedCluster } from "util/clusterLinks";
import { deleteClusterLinkWithTarget } from "util/deleteClusterLink";
import { queryKeys } from "util/queryKeys";
import { pluralize } from "util/helpers";

interface Props {
  cluster: Cluster;
  clusters: Cluster[];
  links: ClusterLink[];
  onDeleted: () => void;
}

const DeleteClusterLinksBtn: FC<Props> = ({
  cluster,
  clusters,
  links,
  onDeleted,
}) => {
  const queryClient = useQueryClient();
  const toastNotification = useToastNotification();
  const [isLoading, setLoading] = useState(false);

  const handleDelete = async () => {
    setLoading(true);
    const failed: string[] = [];
    for (const link of links) {
      const targetCluster = getLinkedCluster(link, clusters);
      const removeTargetEnd =
        link.type === "bidirectional" && !!targetCluster?.tunnel_registered;
      try {
        await deleteClusterLinkWithTarget(
          cluster,
          link,
          targetCluster,
          removeTargetEnd,
        );
      } catch (error) {
        failed.push(link.name);
        toastNotification.failure(
          `Unable to delete cluster link ${link.name}.`,
          error,
        );
      }
    }

    const deletedCount = links.length - failed.length;
    if (deletedCount > 0) {
      toastNotification.success(
        `Deleted ${deletedCount} cluster ${pluralize("link", deletedCount)}.`,
      );
    }

    await queryClient.invalidateQueries({ queryKey: [queryKeys.clusters] });
    onDeleted();
    setLoading(false);
  };

  return (
    <ConfirmationButton
      className="u-no-margin--bottom has-icon"
      loading={isLoading}
      confirmationModalProps={{
        title: "Delete cluster links",
        children: (
          <p>
            This will permanently delete {links.length} cluster{" "}
            {pluralize("link", links.length)}. Bidirectional links are also
            removed from the target cluster when it has a registered tunnel.
            This action cannot be undone.
          </p>
        ),
        confirmButtonLabel: "Delete",
        onConfirm: () => void handleDelete(),
      }}
      shiftClickEnabled
    >
      <Icon name="delete" />
      <span>Delete</span>
    </ConfirmationButton>
  );
};

export default DeleteClusterLinksBtn;

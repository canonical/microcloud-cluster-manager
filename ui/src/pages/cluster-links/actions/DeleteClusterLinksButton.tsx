import {
  ConfirmationButton,
  Icon,
  useNotify,
  useToastNotification,
} from "@canonical/react-components";
import { useQueryClient } from "@tanstack/react-query";
import { deleteClusterLinks } from "api/clusters";
import type { FC } from "react";
import { queryKeys } from "util/queryKeys";
import { pluralize } from "util/helpers";
import type { ClusterConnection } from "util/clusterLinks";

interface Props {
  links: ClusterConnection[];
  onStart: () => void;
  onFinish: () => void;
}

const DeleteClusterLinksButton: FC<Props> = ({ links, onStart, onFinish }) => {
  const queryClient = useQueryClient();
  const notify = useNotify();
  const toastNotification = useToastNotification();

  const handleDelete = () => {
    onStart();
    deleteClusterLinks(
      links.map((entry) => ({
        source: entry.source.name,
        name: entry.link.name,
      })),
    )
      .then(() => {
        toastNotification.success(
          <>
            Deleted{" "}
            <strong>
              {links.length} {pluralize("cluster link", links.length)}
            </strong>
            .
          </>,
        );
      })
      .catch((e) => notify.failure(`Cluster link deletion failed.`, e))
      .finally(() => {
        queryClient.invalidateQueries({
          queryKey: [queryKeys.clusters],
        });
        onFinish();
      });
  };

  return (
    <ConfirmationButton
      appearance=""
      className="p-segmented-control__button u-no-margin--bottom has-icon"
      confirmationModalProps={{
        title: "Confirm delete",
        children: (
          <p>
            Are you sure you want to delete{" "}
            <strong>
              {links.length} {pluralize("cluster link", links.length)}
            </strong>
            ? Only the selected side of each link is removed. Linked clusters
            keep their own cluster links and need to be cleaned up separately.
          </p>
        ),
        confirmButtonLabel: "Delete",
        onConfirm: () => {
          handleDelete();
        },
      }}
      shiftClickEnabled
    >
      <Icon name="delete" />
      <span>Delete {pluralize("cluster link", links.length)}</span>
    </ConfirmationButton>
  );
};

export default DeleteClusterLinksButton;

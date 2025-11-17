import {
  ConfirmationButton,
  Icon,
  useNotify,
  useToastNotification,
} from "@canonical/react-components";
import { useQueryClient } from "@tanstack/react-query";
import { deleteClusterBulk } from "api/clusters";
import type { FC } from "react";
import { queryKeys } from "util/queryKeys";
import { pluralize } from "util/helpers";

interface Props {
  clusterNames: string[];
  onStart: () => void;
  onFinish: () => void;
}

const BulkRemoveClusterButton: FC<Props> = ({
  clusterNames,
  onStart,
  onFinish,
}) => {
  const queryClient = useQueryClient();
  const notify = useNotify();
  const toastNotification = useToastNotification();

  const handleDelete = () => {
    onStart();
    deleteClusterBulk(clusterNames)
      .then(() => {
        toastNotification.success(
          <>
            Removed{" "}
            <strong>
              {clusterNames.length} {pluralize("cluster", clusterNames.length)}
            </strong>
            .
          </>,
        );
      })
      .catch((e) => notify.failure(`Cluster removal failed.`, e))
      .finally(() => {
        void queryClient.invalidateQueries({
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
        title: "Confirm remove",
        children: (
          <>
            <p>
              Are you sure you want to remove{" "}
              <strong>
                {clusterNames.length}{" "}
                {pluralize("cluster", clusterNames.length)}
              </strong>
              ?
            </p>
            <p>
              The {pluralize("cluster", clusterNames.length)} will be be
              unenrolled from cluster manager, but will not be deleted.
            </p>
          </>
        ),
        confirmButtonLabel: "Remove",
        onConfirm: () => {
          handleDelete();
        },
      }}
      shiftClickEnabled
    >
      <Icon name="delete" />
      <span>Remove {pluralize("cluster", clusterNames.length)}</span>
    </ConfirmationButton>
  );
};

export default BulkRemoveClusterButton;

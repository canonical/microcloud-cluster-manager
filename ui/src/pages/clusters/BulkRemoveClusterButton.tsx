import {
  ConfirmationButton,
  Icon,
  useNotify,
} from "@canonical/react-components";
import { useQueryClient } from "@tanstack/react-query";
import { deleteClusterBulk } from "api/clusters";
import { FC } from "react";
import { queryKeys } from "util/queryKeys";
import { pluralize } from "util/helpers";

type Props = {
  clusterNames: string[];
  onStart: () => void;
  onFinish: () => void;
};

const BulkRemoveClusterButton: FC<Props> = ({
  clusterNames,
  onStart,
  onFinish,
}) => {
  const queryClient = useQueryClient();
  const notify = useNotify();

  const handleDelete = () => {
    onStart();
    deleteClusterBulk(clusterNames)
      .then(() => {
        notify.success(
          <>
            {clusterNames.length} {pluralize("cluster", clusterNames.length)}{" "}
            deleted.
          </>,
        );
      })
      .catch((e) => notify.failure(`Cluster deletion failed.`, e))
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
        onConfirm: () => void handleDelete(),
      }}
      shiftClickEnabled
    >
      <Icon name="delete" />
      <span>Remove {pluralize("cluster", clusterNames.length)}</span>
    </ConfirmationButton>
  );
};

export default BulkRemoveClusterButton;

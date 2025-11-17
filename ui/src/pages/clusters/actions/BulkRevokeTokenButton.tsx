import {
  ConfirmationButton,
  Icon,
  useNotify,
  useToastNotification,
} from "@canonical/react-components";
import { useQueryClient } from "@tanstack/react-query";
import type { FC } from "react";
import { queryKeys } from "util/queryKeys";
import { pluralize } from "util/helpers";
import { deleteTokenBulk } from "api/tokens";

interface Props {
  clusterNames: string[];
  onStart: () => void;
  onFinish: () => void;
}

const BulkDeleteClusterButton: FC<Props> = ({
  clusterNames,
  onStart,
  onFinish,
}) => {
  const queryClient = useQueryClient();
  const notify = useNotify();
  const toastNotification = useToastNotification();

  const handleDelete = () => {
    onStart();
    deleteTokenBulk(clusterNames)
      .then(() => {
        toastNotification.success(
          <>
            Revoked{" "}
            <strong>
              {clusterNames.length} {pluralize("token", clusterNames.length)}
            </strong>
            .
          </>,
        );
      })
      .catch((e) => notify.failure(`Token revoke failed.`, e))
      .finally(() => {
        void queryClient.invalidateQueries({
          queryKey: [queryKeys.tokens],
        });
        onFinish();
      });
  };

  return (
    <ConfirmationButton
      appearance=""
      className="p-segmented-control__button u-no-margin--bottom has-icon"
      confirmationModalProps={{
        title: "Confirm revoke",
        children: (
          <>
            <p>
              Are you sure you want to revoke{" "}
              <strong>
                {clusterNames.length} {pluralize("token", clusterNames.length)}
              </strong>
              ?
            </p>
          </>
        ),
        confirmButtonLabel: "Revoke",
        onConfirm: () => {
          handleDelete();
        },
      }}
      shiftClickEnabled
    >
      <Icon name="delete" />
      <span>Revoke {pluralize("token", clusterNames.length)}</span>
    </ConfirmationButton>
  );
};

export default BulkDeleteClusterButton;

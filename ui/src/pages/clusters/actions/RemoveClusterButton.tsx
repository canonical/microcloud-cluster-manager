import {
  ConfirmationButton,
  Icon,
  useNotify,
  useToastNotification,
} from "@canonical/react-components";
import { useQueryClient } from "@tanstack/react-query";
import { deleteCluster } from "api/clusters";
import type { FC } from "react";
import { useState } from "react";
import { queryKeys } from "util/queryKeys";
import { useNavigate } from "react-router-dom";
import classnames from "classnames";

interface Props {
  clusterName: string;
  className?: string;
  onClose?: () => void;
}

const RemoveClusterButton: FC<Props> = ({
  clusterName,
  className,
  onClose,
}) => {
  const queryClient = useQueryClient();
  const notify = useNotify();
  const toastNotification = useToastNotification();
  const navigate = useNavigate();
  const [isLoading, setLoading] = useState(false);

  const handleDelete = async () => {
    setLoading(true);
    try {
      await deleteCluster(clusterName);
      navigate("/ui/clusters");
      await queryClient.invalidateQueries({
        queryKey: [queryKeys.clusters],
      });
      toastNotification.success(
        <>
          Removed cluster <strong>{clusterName}</strong>.
        </>,
      );
    } catch (error) {
      notify.failure(`Unable to remove cluster ${clusterName}.`, error);
    }
    setLoading(false);
    onClose?.();
  };

  return (
    <ConfirmationButton
      className={classnames("u-no-margin--bottom has-icon", className)}
      loading={isLoading}
      confirmationModalProps={{
        title: "Confirm remove",
        children: (
          <>
            <p>
              Are you sure you want to remove the cluster{" "}
              <strong>{clusterName}</strong>?
            </p>
            <p>
              The cluster will be be unenrolled from cluster manager, but it
              will not be deleted.
            </p>
          </>
        ),
        confirmButtonLabel: "Confirm remove",
        onConfirm: () => void handleDelete(),
      }}
      shiftClickEnabled
    >
      <Icon name="delete" />
      <span>Remove</span>
    </ConfirmationButton>
  );
};

export default RemoveClusterButton;

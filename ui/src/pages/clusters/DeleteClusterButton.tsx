import {
  ConfirmationButton,
  Icon,
  useNotify,
} from "@canonical/react-components";
import { useQueryClient } from "@tanstack/react-query";
import { deleteCluster } from "api/clusters";
import { FC, useState } from "react";
import { queryKeys } from "util/queryKeys";
import { useNavigate } from "react-router-dom";

type Props = {
  clusterName: string;
};

const DeleteClusterButton: FC<Props> = ({ clusterName }) => {
  const queryClient = useQueryClient();
  const notify = useNotify();
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
      notify.success(`Successfully deleted cluster ${clusterName}.`);
    } catch (error) {
      notify.failure(`Unable to delete cluster ${clusterName}.`, error);
    }
    setLoading(false);
  };

  return (
    <ConfirmationButton
      appearance=""
      loading={isLoading}
      className="u-no-margin--bottom has-icon"
      confirmationModalProps={{
        title: "Confirm delete",
        children: (
          <p>
            This will permanently delete the cluster{" "}
            <strong>{clusterName}</strong>. This action cannot be undone, and
            can result in data loss.
          </p>
        ),
        confirmButtonLabel: "Delete",
        onConfirm: () => void handleDelete(),
      }}
      shiftClickEnabled
      showShiftClickHint
    >
      <Icon name="delete" />
      <span>Delete</span>
    </ConfirmationButton>
  );
};

export default DeleteClusterButton;

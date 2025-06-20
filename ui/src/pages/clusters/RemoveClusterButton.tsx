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

const RemoveClusterButton: FC<Props> = ({ clusterName }) => {
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
      className="p-segmented-control__button u-no-margin--bottom has-icon"
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

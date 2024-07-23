import { ActionButton, useNotify } from "@canonical/react-components";
import { useQueryClient } from "@tanstack/react-query";
import { deleteCluster } from "api/clusters";
import { FC, useState } from "react";
import { queryKeys } from "util/queryKeys";

type Props = {
  clusterName: string;
};

const DeleteClusterButton: FC<Props> = ({ clusterName }) => {
  const queryClient = useQueryClient();
  const notify = useNotify();
  const [isLoading, setLoading] = useState(false);

  const handleDeleteCluster = async () => {
    setLoading(true);

    try {
      await deleteCluster(clusterName);
      await queryClient.invalidateQueries({
        queryKey: [queryKeys.clusters],
      });
      setLoading(false);
      notify.success(`Successfully deleted cluster ${clusterName}.`);
    } catch (error) {
      setLoading(false);
      notify.failure(`Unable to delete cluster ${clusterName}.`, error);
    }
  };

  return (
    <ActionButton
      className="u-no-margin--bottom"
      onClick={() => void handleDeleteCluster()}
      appearance="negative"
      loading={isLoading}
    >
      Delete
    </ActionButton>
  );
};

export default DeleteClusterButton;

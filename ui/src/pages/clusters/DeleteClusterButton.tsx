import { ActionButton } from "@canonical/react-components";
import { useQueryClient } from "@tanstack/react-query";
import { deleteCluster } from "api/clusters";
import { FC, useState } from "react";
import { queryKeys } from "util/queryKeys";

type Props = {
  clusterName: string;
};

const DeleteClusterButton: FC<Props> = ({ clusterName }) => {
  const queryClient = useQueryClient();
  const [isLoading, setLoading] = useState(false);

  const handleDeleteCluster = async () => {
    setLoading(true);

    try {
      await deleteCluster(clusterName);
      await queryClient.invalidateQueries({
        queryKey: [queryKeys.clusters],
      });
      setLoading(false);
    } catch (error) {
      setLoading(false);
    }
  };

  return (
    <ActionButton
      onClick={() => void handleDeleteCluster()}
      appearance="negative"
      loading={isLoading}
    >
      Delete
    </ActionButton>
  );
};

export default DeleteClusterButton;

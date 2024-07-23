import { ActionButton, useNotify } from "@canonical/react-components";
import { useQueryClient } from "@tanstack/react-query";
import { approveCluster } from "api/clusters";
import { FC, useState } from "react";
import { queryKeys } from "util/queryKeys";

type Props = {
  clusterName: string;
};

const ApproveClusterButton: FC<Props> = ({ clusterName }) => {
  const queryClient = useQueryClient();
  const notify = useNotify();
  const [isLoading, setLoading] = useState(false);

  const handleApproveCluster = async () => {
    setLoading(true);

    try {
      await approveCluster(clusterName);
      await queryClient.invalidateQueries({
        queryKey: [queryKeys.clusters],
      });
      setLoading(false);
      notify.success(`Successfully approved cluster ${clusterName}.`);
    } catch (error) {
      setLoading(false);
      notify.failure(`Unable to approve cluster ${clusterName}.`, error);
    }
  };

  return (
    <ActionButton
      className="u-no-margin--bottom"
      onClick={() => void handleApproveCluster()}
      appearance="positive"
      loading={isLoading}
    >
      Approve
    </ActionButton>
  );
};

export default ApproveClusterButton;

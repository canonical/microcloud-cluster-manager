import { ActionButton } from "@canonical/react-components";
import { useQueryClient } from "@tanstack/react-query";
import { deleteSite } from "api/sites";
import { FC, useState } from "react";
import { queryKeys } from "util/queryKeys";

type Props = {
  siteName: string;
};

const DeleteSiteButton: FC<Props> = ({ siteName }) => {
  const queryClient = useQueryClient();
  const [isLoading, setLoading] = useState(false);

  const handleDeleteSite = async () => {
    setLoading(true);

    try {
      await deleteSite(siteName);
      await queryClient.invalidateQueries({
        queryKey: [queryKeys.sites],
      });
      setLoading(false);
    } catch (error) {
      setLoading(false);
    }
  };

  return (
    <ActionButton
      onClick={() => void handleDeleteSite()}
      appearance="negative"
      loading={isLoading}
    >
      Delete
    </ActionButton>
  );
};

export default DeleteSiteButton;

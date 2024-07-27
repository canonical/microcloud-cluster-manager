import { FC } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { ActionButton, useNotify } from "@canonical/react-components";
import { deleteToken } from "api/tokens";
import { Token } from "types/token";
import { queryKeys } from "util/queryKeys";

interface Props {
  token: Token;
}

const RevokeTokenButton: FC<Props> = ({ token }) => {
  const queryClient = useQueryClient();
  const notify = useNotify();

  const handleDeleteToken = async () => {
    await deleteToken(token.cluster_name)
      .then(() => {
        void queryClient.invalidateQueries({
          queryKey: [queryKeys.tokens],
        });
        notify.success(
          `Successfully deleted token for cluster ${token.cluster_name}.`,
        );
      })
      .catch((e: Error) => {
        notify.failure(`Unable to delete token ${token.cluster_name}.`, e);
      });
  };

  return (
    <ActionButton
      className="u-no-margin"
      appearance="negative"
      onClick={() => void handleDeleteToken()}
    >
      Revoke
    </ActionButton>
  );
};

export default RevokeTokenButton;

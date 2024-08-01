import { FC, useState } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { ConfirmationButton, useNotify } from "@canonical/react-components";
import { deleteToken } from "api/tokens";
import { Token } from "types/token";
import { queryKeys } from "util/queryKeys";

interface Props {
  token: Token;
}

const RevokeTokenButton: FC<Props> = ({ token }) => {
  const queryClient = useQueryClient();
  const notify = useNotify();
  const [loading, setLoading] = useState(false);

  const handleDeleteToken = async () => {
    setLoading(true);
    await deleteToken(token.cluster_name)
      .then(() => {
        void queryClient.invalidateQueries({
          queryKey: [queryKeys.tokens],
        });
        notify.success(
          `Successfully revoked token for cluster ${token.cluster_name}.`,
        );
      })
      .catch((e: Error) => {
        notify.failure(`Unable to revoke token ${token.cluster_name}.`, e);
      })
      .finally(() => {
        setLoading(false);
      });
  };

  return (
    <ConfirmationButton
      appearance="negative"
      loading={loading}
      className="u-no-margin--bottom"
      confirmationModalProps={{
        title: "Confirm revoke",
        children: (
          <p>
            This will permanently revoke the token for cluster{" "}
            <strong>{token.cluster_name}</strong>. This action cannot be undone,
            and can result in data loss.
          </p>
        ),
        confirmButtonLabel: "Revoke",
        onConfirm: () => void handleDeleteToken(),
      }}
      shiftClickEnabled
      showShiftClickHint
    >
      Revoke
    </ConfirmationButton>
  );
};

export default RevokeTokenButton;

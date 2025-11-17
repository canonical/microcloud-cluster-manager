import type { FC } from "react";
import { useState } from "react";
import { useQueryClient } from "@tanstack/react-query";
import {
  ConfirmationButton,
  Icon,
  useNotify,
  useToastNotification,
} from "@canonical/react-components";
import { deleteToken } from "api/tokens";
import type { Token } from "types/token";
import { queryKeys } from "util/queryKeys";

interface Props {
  token: Token;
}

const RevokeTokenButton: FC<Props> = ({ token }) => {
  const queryClient = useQueryClient();
  const notify = useNotify();
  const toastNotification = useToastNotification();
  const [loading, setLoading] = useState(false);

  const handleDeleteToken = async () => {
    setLoading(true);
    await deleteToken(token.cluster_name)
      .then(() => {
        queryClient.invalidateQueries({
          queryKey: [queryKeys.tokens],
        });
        toastNotification.success(
          <>
            Revoked token <strong>{token.cluster_name}</strong>.
          </>,
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
      appearance="base"
      loading={loading}
      className="u-no-margin--bottom has-icon"
      confirmationModalProps={{
        title: "Confirm revoke",
        children: (
          <p>
            Are you sure you want to revoke the token for cluster{" "}
            <strong>{token.cluster_name}</strong>?
          </p>
        ),
        confirmButtonLabel: "Revoke",
        onConfirm: () => void handleDeleteToken(),
      }}
      shiftClickEnabled
    >
      <Icon name="delete" />
    </ConfirmationButton>
  );
};

export default RevokeTokenButton;

import {
  Button,
  Icon,
  useNotify,
  useToastNotification,
} from "@canonical/react-components";
import type { FC } from "react";

interface Props {
  value: string;
  tooltipMessage: string;
}

const CopyToClipboard: FC<Props> = ({ value, tooltipMessage }) => {
  const notify = useNotify();
  const toastNotification = useToastNotification();

  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(value);
      toastNotification.success(<>Copied to clipboard.</>);
    } catch (error) {
      notify.failure("Copying to clipboard failed.", error);
    }
  };

  return (
    <Button
      appearance=""
      aria-label={tooltipMessage}
      className="u-no-margin--bottom has-icon"
      onClick={() => void handleCopy()}
      title={tooltipMessage}
      type="button"
      hasIcon
    >
      <Icon name="copy" />
    </Button>
  );
};

export default CopyToClipboard;

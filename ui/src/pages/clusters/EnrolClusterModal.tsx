import { FC, useEffect } from "react";
import { useState } from "react";
import {
  Button,
  CodeSnippet,
  CodeSnippetBlockAppearance,
  Icon,
  Modal,
} from "@canonical/react-components";
import { convertToISOFormat, isoTimeToString } from "util/helpers";

interface Props {
  onClose: () => void;
  token: string;
  name: string;
  expiry: string;
}

const EnrolClusterModal: FC<Props> = ({ onClose, token, name, expiry }) => {
  const [copied, setCopied] = useState(false);
  const command = `microcloud cluster-manager join ${token}`;

  useEffect(() => {
    window.history.replaceState(null, "", location.pathname); // clear the location state
  }, []);

  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(command);
      setCopied(true);

      setTimeout(() => {
        setCopied(false);
      }, 5000);
    } catch (error) {
      console.error(error);
    }
  };

  return (
    <Modal
      close={onClose}
      closeOnOutsideClick={false}
      className="u-modal-medium"
      title={`Join token for cluster ${name} created`}
      buttonRow={
        <>
          {token && (
            <>
              <Button
                aria-label={
                  copied ? "Copied to clipboard" : "Copy to clipboard"
                }
                title="Copy token"
                className="u-no-margin--bottom"
                onClick={handleCopy}
                type="button"
                hasIcon
              >
                <Icon name={copied ? "task-outstanding" : "copy"} />
                <span>Copy command</span>
              </Button>
              <Button
                aria-label="Close"
                className="u-no-margin--bottom"
                onClick={onClose}
                type="button"
              >
                Close
              </Button>
            </>
          )}
        </>
      }
    >
      <>
        <p>
          To finish the enrollment, run the command below on any member of the
          MicroCloud. The command is valid until{" "}
          {isoTimeToString(convertToISOFormat(expiry))}.
        </p>
        <CodeSnippet
          className="adb-connect-wrapper"
          blocks={[
            {
              appearance: CodeSnippetBlockAppearance.LINUX_PROMPT,
              code: (
                <>
                  <span className="command u-truncate" title={command}>
                    {command}
                  </span>
                </>
              ),
            },
          ]}
        />
        <p>
          <b>
            Once this modal is closed, the command can&rsquo;t be viewed again.
          </b>
        </p>
      </>
    </Modal>
  );
};

export default EnrolClusterModal;

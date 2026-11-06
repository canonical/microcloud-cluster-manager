import { Button, Icon } from "@canonical/react-components";
import type { FC } from "react";
import classnames from "classnames";

interface Props {
  // Without an href the button is rendered disabled.
  href?: string;
  label: string;
  iconName: string;
  unavailableReason: string;
  className?: string;
  onClose?: () => void;
  hasLabel?: boolean;
}

// A link styled as a button that opens in a new tab. Being an anchor keeps
// native link behaviour (middle click, copy link address).
const ExternalLinkButton: FC<Props> = ({
  href,
  label,
  iconName,
  unavailableReason,
  className,
  onClose,
  hasLabel = true,
}) => {
  const isAvailable = !!href;

  return (
    <Button
      element="a"
      appearance={hasLabel ? undefined : "base"}
      hasIcon
      className={classnames("u-no-margin--bottom", className)}
      disabled={!isAvailable}
      href={href}
      target={isAvailable ? "_blank" : undefined}
      rel={isAvailable ? "noopener noreferrer" : undefined}
      onClick={isAvailable ? onClose : undefined}
      aria-label={
        hasLabel ? undefined : `${label}${isAvailable ? "" : " (unavailable)"}`
      }
      title={isAvailable ? (hasLabel ? undefined : label) : unavailableReason}
    >
      <Icon name={iconName} />
      {hasLabel && <span>{label}</span>}
    </Button>
  );
};

export default ExternalLinkButton;

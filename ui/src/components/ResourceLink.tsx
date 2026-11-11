import type { FC } from "react";
import { Link } from "react-router-dom";
import ResourceIcon, { type ResourceIconType } from "./ResourceIcon";
import classnames from "classnames";

interface Props {
  type: ResourceIconType;
  value: string;
  to?: string;
  disabled?: boolean;
  hasTitle?: boolean;
  className?: string;
}

const ResourceLink: FC<Props> = ({
  type,
  value,
  to,
  disabled,
  hasTitle = true,
  className,
}) => {
  const commonProps = {
    className: classnames(
      "p-chip is-inline is-dense resource-link",
      {
        "p-chip--disabled": disabled,
      },
      className,
    ),
    title: hasTitle ? value : undefined,
  };

  const content = (
    <>
      <ResourceIcon type={type} />
      <span className="p-chip__value">{value}</span>
    </>
  );

  if (to) {
    return (
      <Link to={to} {...commonProps}>
        {content}
      </Link>
    );
  }

  return <span {...commonProps}>{content}</span>;
};

export default ResourceLink;

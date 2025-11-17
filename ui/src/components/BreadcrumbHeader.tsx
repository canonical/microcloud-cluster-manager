import { Tooltip } from "@canonical/react-components";
import type { FC, ReactNode } from "react";

interface Props {
  name: string;
  parentItems?: ReactNode[];
}

const BreadCrumbHeader: FC<Props> = ({ name, parentItems }: Props) => {
  return (
    <nav
      className="p-breadcrumbs p-breadcrumbs--large"
      aria-label="Breadcrumbs"
    >
      <ol className="p-breadcrumbs__items">
        {parentItems
          ? parentItems.map((item, key) => (
              <li
                className="p-heading--4 u-no-margin--bottom continuous-breadcrumb"
                key={key}
              >
                {item}
              </li>
            ))
          : null}

        <li
          className="p-heading--4 u-no-margin--bottom name continuous-breadcrumb"
          title={name}
        >
          <Tooltip message={name} position="btm-left">
            {name}
          </Tooltip>
        </li>
      </ol>
    </nav>
  );
};

export default BreadCrumbHeader;

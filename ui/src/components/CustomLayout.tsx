import { FC, ReactNode } from "react";
import classnames from "classnames";
import { AppMain } from "@canonical/react-components";

interface Props {
  header?: ReactNode;
  children: ReactNode;
  mainClassName?: string;
  contentClassName?: string;
}

const CustomLayout: FC<Props> = ({
  header,
  children,
  contentClassName,
  mainClassName,
}: Props) => {
  return (
    <AppMain className={mainClassName}>
      <div className="p-panel">
        {header}
        <div className={classnames("p-panel__content", contentClassName)}>
          {children}
        </div>
      </div>
    </AppMain>
  );
};

export default CustomLayout;

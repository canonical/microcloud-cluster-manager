import type { ElementType, FC, ReactNode } from "react";
import { AppMain, Panel } from "@canonical/react-components";

interface Props {
  title: string | ReactNode;
  controls?: ReactNode;
  children: ReactNode;
  mainClassName?: string;
  contentClassName?: string;
  titleClassName?: string;
  titleComponent?: ElementType;
}

const BaseLayout: FC<Props> = ({
  title,
  controls,
  children,
  mainClassName,
  contentClassName,
  titleClassName,
  titleComponent,
}: Props) => {
  return (
    <AppMain className={mainClassName}>
      <Panel
        title={title}
        controls={controls}
        wrapContent={true}
        contentClassName={contentClassName}
        titleClassName={titleClassName}
        titleComponent={titleComponent}
      >
        {children}
      </Panel>
    </AppMain>
  );
};

export default BaseLayout;

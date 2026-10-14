import type { FC, ReactNode } from "react";
import { List } from "@canonical/react-components";

interface Props {
  items: ReactNode[];
}

const ActionsList: FC<Props> = ({ items }) => {
  return (
    <List inline className="actions-list u-no-margin--bottom" items={items} />
  );
};

export default ActionsList;

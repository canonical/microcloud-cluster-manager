import type { FC } from "react";
import editOffIcon from "icons/edit_off.svg";
import emptyIcon from "icons/empty.svg";

const ICONS = {
  "edit-off": editOffIcon,
  empty: emptyIcon,
};

export type CustomIconName = keyof typeof ICONS;

interface Props {
  name: CustomIconName;
  size?: number;
  className?: string;
}

const CustomIcon: FC<Props> = ({ name, size, className }) => {
  return (
    <img
      src={ICONS[name]}
      alt=""
      width={size}
      height={size}
      className={className}
    />
  );
};

export default CustomIcon;

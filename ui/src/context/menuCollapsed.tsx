import { useState } from "react";
import useEventListener from "@use-it/event-listener";
import { isWidthBelow } from "util/helpers";

const isSmallScreen = () => isWidthBelow(620);
const isMediumScreen = () => isWidthBelow(820);

export const useMenuCollapsed = () => {
  const [menuCollapsed, setMenuCollapsed] = useState(isMediumScreen());

  const collapseOnMediumScreen = () => {
    if (isSmallScreen()) {
      return;
    }
  };

  useEventListener("resize", collapseOnMediumScreen);

  return { menuCollapsed, setMenuCollapsed };
};

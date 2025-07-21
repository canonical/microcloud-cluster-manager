import { useState } from "react";
import { isWidthBelow } from "util/helpers";
import { useListener } from "@canonical/react-components";

const isSmallScreen = () => isWidthBelow(620);
const isMediumScreen = () => isWidthBelow(820);

export const useMenuCollapsed = () => {
  const [menuCollapsed, setMenuCollapsed] = useState(isMediumScreen());

  const collapseOnMediumScreen = () => {
    if (isSmallScreen()) {
      return;
    }

    setMenuCollapsed(isMediumScreen());
  };

  useListener(window, collapseOnMediumScreen, "resize", true);

  return { menuCollapsed, setMenuCollapsed };
};

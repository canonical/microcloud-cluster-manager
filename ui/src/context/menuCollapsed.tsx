import { useState } from "react";
import { isDimensionBelow } from "util/helpers";
import { useListener } from "@canonical/react-components";

const isSmallScreen = () => isDimensionBelow(620);
const isMediumScreen = () => isDimensionBelow(820);

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

import { Icon } from "@canonical/react-components";
import {
  useIsScreenBelow,
  largeScreenBreakpoint,
} from "context/useIsScreenBelow";
import type { FC } from "react";

const NotAuthorized: FC = () => {
  const isMediumScreen = useIsScreenBelow(largeScreenBreakpoint, "width");
  return (
    <div className="not-authorized">
      <Icon name="warning" className="not-authorized__icon u-hide--small" />
      <span>
        {isMediumScreen
          ? "Limited access. Contact admin."
          : "You are not authorized to access cluster manager. Please contact your administrator to give you admin access."}
      </span>
    </div>
  );
};

export default NotAuthorized;

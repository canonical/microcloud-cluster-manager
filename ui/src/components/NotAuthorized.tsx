import { Icon } from "@canonical/react-components";
import type { FC } from "react";

const NotAuthorized: FC = () => {
  return (
    <div className="not-authorized">
      <Icon name="warning" />
      <span>
        You are not authorized to access cluster manager. Please contact your
        administrator to give you admin access.
      </span>
    </div>
  );
};

export default NotAuthorized;

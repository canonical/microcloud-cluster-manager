import type { FC } from "react";
import { NavLink } from "react-router-dom";

const Logo: FC = () => {
  return (
    <NavLink className="p-panel__logo" to={`/`}>
      <img
        src="/ui/assets/img/lxd-logo.svg"
        alt="Cluster manager logo"
        className="p-panel__logo-image"
      />
      <div className="logo-text p-heading--4">Cluster Manager</div>
    </NavLink>
  );
};

export default Logo;

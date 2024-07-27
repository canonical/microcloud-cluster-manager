import { FC, MouseEvent } from "react";
import {
  AppNavigation,
  AppNavigationBar,
  Button,
  Icon,
  Panel,
} from "@canonical/react-components";
import classnames from "classnames";
import Logo from "./Logo";
import NavLink from "components/NavLink";
import { useMenuCollapsed } from "context/menuCollapsed";
import { isWidthBelow, logout } from "util/helpers";
import { useAuth } from "context/auth";

const Navigation: FC = () => {
  const isSmallScreen = () => isWidthBelow(620);
  const { menuCollapsed, setMenuCollapsed } = useMenuCollapsed();
  const { isAuthenticated } = useAuth();

  const handleLogout = () => {
    logout();
    softToggleMenu();
  };

  const softToggleMenu = () => {
    if (isSmallScreen()) {
      setMenuCollapsed((prev) => !prev);
    }
  };

  const hardToggleMenu = (e: MouseEvent<HTMLElement>) => {
    setMenuCollapsed((prev) => !prev);
    e.stopPropagation();
  };

  if (!isAuthenticated) {
    return null;
  }

  return (
    <>
      <AppNavigationBar>
        <Panel
          className="is-dark"
          stickyHeader={true}
          logo={<Logo />}
          controls={
            <Button dense className="p-panel__toggle" onClick={hardToggleMenu}>
              Menu
            </Button>
          }
        />
      </AppNavigationBar>
      <AppNavigation
        aria-label="main navigation"
        className={classnames({
          "is-collapsed": menuCollapsed,
          "is-pinned": !menuCollapsed,
        })}
      >
        <Panel
          logo={<Logo />}
          dark={true}
          controls={
            <Button
              appearance="base"
              hasIcon
              className="u-no-margin"
              aria-label="close navigation"
              onClick={hardToggleMenu}
            >
              <Icon name="close" />
            </Button>
          }
          controlsClassName="u-hide--medium u-hide--large"
        >
          <div className="p-side-navigation--icons is-dark">
            <ul className="p-side-navigation__list sidenav-top-ul">
              <li>
                <NavLink
                  to={`/ui/clusters`}
                  title={`Clusters List`}
                  onClick={softToggleMenu}
                >
                  <img
                    src="/ui/assets/img/cluster-icon.svg"
                    alt="cluster-icon"
                    className="p-side-navigation__icon"
                  />
                  Clusters
                </NavLink>
              </li>
              <li>
                <NavLink
                  to={`/ui/settings`}
                  title={`Settings`}
                  onClick={softToggleMenu}
                >
                  <Icon
                    className="is-light p-side-navigation__icon"
                    name="settings"
                  />
                  Settings
                </NavLink>
              </li>
            </ul>
            <ul className="p-side-navigation__list sidenav-bottom-ul">
              <hr className="is-dark navigation-hr" />
              <li className="p-side-navigation__item">
                <a
                  className="p-side-navigation__link"
                  title="Log out"
                  onClick={handleLogout}
                >
                  <Icon
                    className="is-light p-side-navigation__icon p-side-logout"
                    name="export"
                  />
                  Log out
                </a>
              </li>
            </ul>
          </div>
        </Panel>

        <div className="sidenav-toggle-wrapper">
          <Button
            appearance="base"
            aria-label={`${
              menuCollapsed ? "expand" : "collapse"
            } main navigation`}
            hasIcon
            dense
            className="sidenav-toggle is-dark u-no-margin l-navigation-collapse-toggle u-hide--small"
            onClick={hardToggleMenu}
          >
            <Icon light name="sidebar-toggle" />
          </Button>
        </div>
      </AppNavigation>
    </>
  );
};

export default Navigation;

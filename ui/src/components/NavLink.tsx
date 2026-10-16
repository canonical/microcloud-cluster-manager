import type { FC, LinkHTMLAttributes, ReactNode } from "react";
import { Link, matchPath, useLocation } from "react-router-dom";
import classnames from "classnames";

interface Props {
  to: string;
  title: string;
  children: ReactNode;
  className?: string;
  activeUrlMatches?: string[];
  ignoreUrlMatches?: string[];
}

const NavLink: FC<Props & LinkHTMLAttributes<HTMLElement>> = ({
  to,
  title,
  children,
  className,
  activeUrlMatches = [],
  ignoreUrlMatches = [],
  ...linkProps
}) => {
  const location = useLocation();

  // ignore last char to consider plural form of URL Paths e.g. /clusters and /cluster/1
  const matchPart = to.substring(0, to.length - 1);
  let isActive = location.pathname.startsWith(matchPart);

  // this caters for more custom url matches for link to be active.
  for (const match of activeUrlMatches) {
    if (
      location.pathname.includes(match) ||
      matchPath({ path: match, end: false }, location.pathname)
    ) {
      isActive = true;
    }
  }

  // this caters for url matches that should not activate the link,
  // e.g. /ui/cluster-links should not activate /ui/clusters
  for (const match of ignoreUrlMatches) {
    if (location.pathname.includes(match)) {
      isActive = false;
    }
  }

  return (
    <Link
      title={title}
      to={to}
      aria-current={isActive ? "page" : undefined}
      className={classnames("p-side-navigation__link", className)}
      {...linkProps}
    >
      {children}
    </Link>
  );
};

export default NavLink;

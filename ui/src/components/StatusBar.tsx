import type { FC } from "react";
import {
  Button,
  Icon,
  ICONS,
  useListener,
  useToastNotification,
} from "@canonical/react-components";
import { useQuery } from "@tanstack/react-query";
import { queryKeys } from "util/queryKeys";
import { fetchConfigurations } from "api/settings";
import {
  iconLookup,
  severityOrder,
} from "@canonical/react-components/dist/components/Notifications/ToastNotification/ToastNotificationList";
import classNames from "classnames";
import NotAuthorized from "./NotAuthorized";
import { useAuth } from "context/auth";

interface Props {
  className?: string;
}

const StatusBar: FC<Props> = () => {
  const { data: configurations } = useQuery({
    queryKey: [queryKeys.configuration],
    queryFn: fetchConfigurations,
  });

  const version = configurations?.version?.value;

  const { toggleListView, notifications, countBySeverity, isListView } =
    useToastNotification();

  const { isAdmin, isAuthenticated } = useAuth();

  useListener(
    window,
    (e: KeyboardEvent) => {
      // Close notifications list if Escape pressed
      if (e.code === "Escape" && isListView) {
        toggleListView();
      }
    },
    "keydown",
  );

  const notificationIcons = severityOrder.map((severity) => {
    if (countBySeverity[severity]) {
      return (
        <Icon
          key={severity}
          name={iconLookup[severity]}
          aria-label={`${severity} notification exists`}
        />
      );
    }
    return null;
  });

  const hasNotifications = notifications.length > 0;

  return (
    <div className="l-status status-bar" id="status-bar">
      <span className="server-version p-text--small">Version {version}</span>
      {isAuthenticated && !isAdmin && <NotAuthorized />}
      {hasNotifications && (
        <Button
          className={classNames("u-no-margin expand-button", {
            "button-active": isListView,
          })}
          onClick={toggleListView}
          aria-label="Expand notifications list"
        >
          {notificationIcons}
          <span className="total-count">{notifications.length}</span>
          <Icon name={isListView ? ICONS.chevronDown : ICONS.chevronUp} />
        </Button>
      )}
    </div>
  );
};

export default StatusBar;

import { FC } from "react";
import {
  Button,
  Icon,
  ICONS,
  useToastNotification,
} from "@canonical/react-components";
import { useQuery } from "@tanstack/react-query";
import { queryKeys } from "util/queryKeys";
import { fetchConfigurations } from "api/settings";
import useEventListener from "@use-it/event-listener";
import {
  iconLookup,
  severityOrder,
} from "@canonical/react-components/dist/components/Notifications/ToastNotification/ToastNotificationList";
import classNames from "classnames";

interface Props {
  className?: string;
}

const StatusBar: FC<Props> = () => {
  const { data: configurations } = useQuery({
    queryKey: [queryKeys.configuration],
    queryFn: fetchConfigurations,
  });

  const version = configurations?.api_version?.value;

  const { toggleListView, notifications, countBySeverity, isListView } =
    useToastNotification();

  useEventListener("keydown", (e: KeyboardEvent) => {
    // Close notifications list if Escape pressed
    if (e.code === "Escape" && isListView) {
      toggleListView();
    }
  });

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

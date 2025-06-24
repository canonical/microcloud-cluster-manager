import { FC } from "react";
import App from "./App";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import {
  Application,
  NotificationProvider,
  QueuedNotification,
  ToastNotificationProvider,
} from "@canonical/react-components";
import Navigation from "./components/Navigation";
import { AuthProvider } from "context/auth";
import StatusBar from "components/StatusBar";
import { useLocation } from "react-router-dom";
import { FetchError } from "util/helpers";

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: (failureCount, error) => {
        // Disable retries for 403 errors
        if (error instanceof FetchError && error.response.error_code === 403) {
          return false;
        }
        // Retry other errors up to 3 times
        return failureCount < 3;
      },
    },
  },
});

const Root: FC = () => {
  const location = useLocation() as QueuedNotification;

  return (
    <QueryClientProvider client={queryClient}>
      <AuthProvider>
        <Application id="l-application">
          <ToastNotificationProvider>
            <NotificationProvider
              state={location.state}
              pathname={location.pathname}
            >
              <Navigation />
              <App />
              <StatusBar />
            </NotificationProvider>
          </ToastNotificationProvider>
        </Application>
      </AuthProvider>
    </QueryClientProvider>
  );
};

export default Root;

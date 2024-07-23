import { FC } from "react";
import App from "./App";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import {
  Application,
  NotificationProvider,
  QueuedNotification,
} from "@canonical/react-components";
import Navigation from "./components/Navigation";
import { AuthProvider } from "context/auth";
import StatusBar from "components/StatusBar";
import { useLocation } from "react-router-dom";

const queryClient = new QueryClient();

const Root: FC = () => {
  const location = useLocation() as QueuedNotification;

  return (
    <QueryClientProvider client={queryClient}>
      <AuthProvider>
        <Application>
          <NotificationProvider
            state={location.state}
            pathname={location.pathname}
          >
            <Navigation />
            <App />
            <StatusBar />
          </NotificationProvider>
        </Application>
      </AuthProvider>
    </QueryClientProvider>
  );
};

export default Root;

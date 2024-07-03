import { FC } from "react";
import App from "./App";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { Application } from "@canonical/react-components";

import Navigation from "./components/Navigation";

const queryClient = new QueryClient();

const Root: FC = () => {
  return (
    <QueryClientProvider client={queryClient}>
      <Application>
        <Navigation />
        <App />
      </Application>
    </QueryClientProvider>
  );
};

export default Root;

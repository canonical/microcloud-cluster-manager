import React, { FC, lazy, Suspense } from "react";
import { Navigate, Route, Routes, useLocation } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import { queryKeys } from "util/queryKeys";
import { fetchSites } from "api/sites";
import NoMatch from "pages/NoMatch";

const SiteList = lazy(() => import("pages/sites/SiteList"));
const Login = lazy(() => import("pages/Login"));

const App: FC = () => {
  const { pathname } = useLocation();

  const { error, isLoading } = useQuery({
    queryKey: [queryKeys.sites],
    queryFn: fetchSites,
    retry: false,
  });

  const isAuthError = error?.message === "not authorized";
  const isLoginPath = pathname === "/ui/login";

  if (!isLoading && !isLoginPath && isAuthError) {
    window.location.href = "/ui/login";
    return null;
  }

  if (!isLoading && isLoginPath && !isAuthError) {
    window.location.href = "/ui/sites";
    return null;
  }

  return (
    <Suspense fallback={<div>Loading</div>}>
      <Routes>
        <Route path="/" element={<Navigate to="/ui/sites" replace={true} />} />
        <Route
          path="/ui"
          element={<Navigate to="/ui/sites" replace={true} />}
        />
        <Route path="/ui/login" element={<Login />} />
        <Route path="/ui/sites" element={<SiteList />} />
        <Route path="*" element={<NoMatch />} />
      </Routes>
    </Suspense>
  );
};

export default App;

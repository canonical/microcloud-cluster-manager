import { FC, lazy, Suspense } from "react";
import { Navigate, Route, Routes } from "react-router-dom";
import NoMatch from "pages/NoMatch";
import Settings from "pages/Settings";
import { useAuth } from "context/auth";
import CreateCluster from "pages/clusters/CreateCluster";
import { logout } from "util/helpers";

const ClusterList = lazy(() => import("pages/clusters/ClusterList"));
const Login = lazy(() => import("pages/Login"));

const App: FC = () => {
  const { isAuthLoading, isAuthenticated } = useAuth();

  const preLoginRoutes = (
    <>
      <Route path="/ui/login" element={<Login />} />
      <Route path="*" element={<Navigate to="/ui/login" replace={true} />} />
    </>
  );

  const loggedInRoutes = (
    <>
      <Route path="/" element={<Navigate to="/ui/clusters" replace={true} />} />
      <Route
        path="/ui"
        element={<Navigate to="/ui/clusters" replace={true} />}
      />
      <Route path="/ui/clusters" element={<ClusterList />} />
      <Route path="/ui/clusters/:activeTab" element={<ClusterList />} />
      <Route path="/ui/clusters/create" element={<CreateCluster />} />
      <Route path="/ui/settings" element={<Settings />} />
      <Route path="*" element={<NoMatch />} />
    </>
  );

  if (isAuthLoading) {
    return;
  }

  if (!isAuthenticated) {
    logout();
  }

  return (
    <Suspense>
      <Routes>{isAuthenticated ? loggedInRoutes : preLoginRoutes}</Routes>
    </Suspense>
  );
};

export default App;

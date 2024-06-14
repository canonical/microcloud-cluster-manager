import { FC, lazy, Suspense } from "react";
import { Navigate, Route, Routes } from "react-router-dom";

const SiteList = lazy(() => import("pages/sites/SiteList"));

const App: FC = () => {
  return (
    <Suspense fallback={<div>Loading</div>}>
      <Routes>
        <Route path="/" element={<Navigate to="/ui" replace={true} />} />
        <Route path="/ui" element={<SiteList />} />
      </Routes>
    </Suspense>
  );
};

export default App;

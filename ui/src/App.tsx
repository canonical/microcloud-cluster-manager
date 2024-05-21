import { FC, lazy, Suspense } from "react";
import { Route, Routes } from "react-router-dom";

const SiteList = lazy(() => import("pages/sites/SiteList"));

const App: FC = () => {
  return (
    <Suspense fallback={<div>Loading</div>}>
      <Routes>
        <Route path="/" element={<SiteList />} />
      </Routes>
    </Suspense>
  );
};

export default App;

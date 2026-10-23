import type { FC } from "react";
import { Row } from "@canonical/react-components";
import { Link } from "react-router-dom";
import BaseLayout from "components/BaseLayout";

const NoMatch: FC = () => {
  return (
    <BaseLayout title="404 Page not found">
      <Row>
        <p>Sorry, we cannot find the page that you are looking for.</p>
        <Link to="/ui/clusters">Go to index page</Link>
      </Row>
    </BaseLayout>
  );
};

export default NoMatch;

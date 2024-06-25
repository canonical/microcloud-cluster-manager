import { FC } from "react";
import { Link, LoginPageLayout } from "@canonical/react-components";

const NoMatch: FC = () => {
  return (
    <LoginPageLayout title="404 Page not found">
      Sorry, we cannot find the page that you are looking for.
      <br />
      <Link href="/">Go to index page</Link>
    </LoginPageLayout>
  );
};

export default NoMatch;

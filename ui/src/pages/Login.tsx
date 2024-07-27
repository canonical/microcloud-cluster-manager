import { FC } from "react";
import { Link, LoginPageLayout } from "@canonical/react-components";
import BaseLayout from "components/BaseLayout";

const Login: FC = () => {
  const postLoginPath = "/ui";

  return (
    <BaseLayout title="">
      <LoginPageLayout title="Login to LXD Cluster Manager">
        <Link href={`/oidc/login?next=${postLoginPath}`} className="p-button">
          Login
        </Link>
      </LoginPageLayout>
    </BaseLayout>
  );
};

export default Login;

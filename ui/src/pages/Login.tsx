import { FC } from "react";
import { Link, LoginPageLayout } from "@canonical/react-components";
import BaseLayout from "components/BaseLayout";

const Login: FC = () => {
  const postLoginPath = "/ui";

  return (
    <BaseLayout title="">
      <LoginPageLayout title="MicroCloud Cluster Manager">
        <p>Access your dashboard to manage MicroCloud clusters</p>
        <Link
          href={`/oidc/login?next=${postLoginPath}`}
          className="p-button--positive"
        >
          Login
        </Link>
      </LoginPageLayout>
    </BaseLayout>
  );
};

export default Login;

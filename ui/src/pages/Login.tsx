import { FC } from "react";
import { Link, LoginPageLayout } from "@canonical/react-components";

const Login: FC = () => {
  return (
    <LoginPageLayout title="Login to LXD site manager">
      <Link href="/oidc/sites" className="p-button">
        Login
      </Link>
    </LoginPageLayout>
  );
};

export default Login;

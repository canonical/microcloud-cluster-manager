import { FC } from "react";
import { Button, LoginPageLayout } from "@canonical/react-components";
import { useNavigate } from "react-router-dom";

const Login: FC = () => {
  const navigate = useNavigate();

  return (
    <LoginPageLayout title="Login to LXD site manager">
      <Button onClick={() => navigate("/oidc/login")}>Login</Button>
    </LoginPageLayout>
  );
};

export default Login;

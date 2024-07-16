import { ActionButton } from "@canonical/react-components";
import { FC } from "react";
import { useNavigate } from "react-router-dom";

const AddClusterButton: FC = () => {
  const navigate = useNavigate();

  return (
    <ActionButton
      onClick={() => {
        navigate("/");
      }}
      appearance="positive"
    >
      Add New Cluster
    </ActionButton>
  );
};

export default AddClusterButton;

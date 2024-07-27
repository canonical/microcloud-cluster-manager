import { ActionButton } from "@canonical/react-components";
import { FC } from "react";
import { useNavigate } from "react-router-dom";

const AddClusterButton: FC = () => {
  const navigate = useNavigate();

  return (
    <ActionButton
      onClick={() => {
        navigate("/ui/clusters/create");
      }}
      appearance="positive"
      className="u-no-margin--bottom"
    >
      Add New Cluster
    </ActionButton>
  );
};

export default AddClusterButton;

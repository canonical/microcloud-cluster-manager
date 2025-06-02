import { ActionButton, Icon } from "@canonical/react-components";
import { FC } from "react";
import { useNavigate } from "react-router-dom";

const AddClusterButton: FC = () => {
  const navigate = useNavigate();

  return (
    <ActionButton
      onClick={() => {
        void navigate("/ui/clusters/create");
      }}
      appearance="positive"
      className="u-float-right u-no-margin--bottom has-icon"
    >
      <Icon name="plus" light />
      <span>Enrol cluster</span>
    </ActionButton>
  );
};

export default AddClusterButton;

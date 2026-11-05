import { Input } from "@canonical/react-components";
import type { FC } from "react";
import CustomIcon from "components/CustomIcon";

interface Props {
  id: string;
  label: string;
  value: string;
  placeholder?: string;
}

const ReadOnlyInput: FC<Props> = ({ id, label, value, placeholder }) => (
  <Input
    id={id}
    type="text"
    className="read-only-input"
    label={
      <>
        {label} <CustomIcon name="edit-off" size={16} />
      </>
    }
    value={value}
    placeholder={placeholder}
    readOnly
  />
);

export default ReadOnlyInput;

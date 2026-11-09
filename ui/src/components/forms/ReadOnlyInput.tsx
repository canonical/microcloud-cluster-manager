import { Input } from "@canonical/react-components";
import type { FC } from "react";
import CustomIcon from "components/CustomIcon";

interface Props {
  label: string;
  value: string;
  placeholder?: string;
}

const ReadOnlyInput: FC<Props> = ({ label, value, placeholder }) => (
  <Input
    type="text"
    className="read-only-input"
    labelClassName="read-only-label"
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

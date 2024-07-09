import { FC, useState } from "react";
import { Button, Form, Input } from "@canonical/react-components";

interface Props {
  initialValue: string;
  configField: string;
  onSubmit: (newValue: string | boolean) => void;
  onCancel: () => void;
}

const getConfigId = (key: string) => {
  return key.replace(".", "___");
};

const SettingFormInput: FC<Props> = ({
  initialValue,
  configField,
  onSubmit,
  onCancel,
}) => {
  const [value, setValue] = useState(initialValue);

  const getInputType = () => {
    return "text";
  };

  return (
    <Form
      onSubmit={(e) => {
        e.preventDefault();
        onSubmit(value);
      }}
    >
      {/* hidden submit to enable enter key in inputs */}
      <Input type="submit" hidden value="Hidden input" />
      <Input
        aria-label={configField}
        id={getConfigId(configField)}
        wrapperClassName="input-wrapper"
        type={getInputType()}
        value={String(value)}
        onChange={(e) => setValue(e.target.value)}
      />
      <Button appearance="base" onClick={onCancel}>
        Cancel
      </Button>
      <Button appearance="positive" onClick={() => onSubmit(value)}>
        Save
      </Button>
    </Form>
  );
};

export default SettingFormInput;

import { SearchBox } from "@canonical/react-components";
import { useId, type FC } from "react";

interface Props {
  onChange: (val: string) => void;
  value: string;
  disabled?: boolean;
  className?: string;
  labelledBy?: string;
}

const PermissionGroupsFilter: FC<Props> = ({
  onChange,
  value,
  disabled,
  className,
  labelledBy,
}) => {
  const id = useId();
  const handleSearchChange = (value: string) => {
    onChange(value.toLowerCase());
  };

  return (
    <div className="permission-groups-filter">
      <SearchBox
        id={id}
        aria-labelledby={labelledBy}
        label="Search groups"
        name={id}
        type="text"
        onSearch={handleSearchChange}
        onChange={handleSearchChange}
        value={value}
        placeholder="Search groups"
        disabled={disabled}
        className={className}
      />
    </div>
  );
};

export default PermissionGroupsFilter;

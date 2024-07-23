import { FC, useEffect, useRef, useState } from "react";
import { Button, Icon, useNotify } from "@canonical/react-components";
import { updateManagerConfigs, updateMemberConfigs } from "api/settings";
import { queryKeys } from "util/queryKeys";
import { useQueryClient } from "@tanstack/react-query";
import SettingFormInput from "./SettingFormInput";

interface Props {
  configField: string;
  value?: string;
  isLast?: boolean;
  member?: string;
}

const SettingForm: FC<Props> = ({ configField, value, isLast, member }) => {
  const [isEditMode, setEditMode] = useState(false);
  const notify = useNotify();
  const queryClient = useQueryClient();

  const editRef = useRef<HTMLDivElement | null>(null);

  const onSubmit = (newValue: string | boolean) => {
    const config = {
      [configField]: String(newValue),
    };

    (member
      ? updateMemberConfigs(member, config)
      : updateManagerConfigs(config)
    )
      .then(() => {
        setEditMode(false);
        notify.success(`Setting ${configField} updated.`);
      })
      .catch((e) => {
        notify.failure("Setting update failed", e);
      })
      .finally(() => {
        void queryClient.invalidateQueries({
          queryKey: [
            member
              ? queryKeys.memberConfigOptions
              : queryKeys.managerConfigOptions,
          ],
        });
      });
  };

  const onCancel = () => {
    setEditMode(false);
  };

  const getReadModeValue = () => {
    return value ? value : "-";
  };

  useEffect(() => {
    if (isEditMode && isLast) {
      editRef.current?.scrollIntoView({ behavior: "smooth" });
    }
  }, [isEditMode]);

  return (
    <>
      {isEditMode && (
        <div ref={editRef}>
          <SettingFormInput
            initialValue={value ?? ""}
            configField={configField}
            onSubmit={onSubmit}
            onCancel={onCancel}
          />
        </div>
      )}
      {!isEditMode && (
        <>
          <Button
            appearance="base"
            className="readmode-button u-no-margin"
            onClick={() => {
              setEditMode(true);
            }}
            hasIcon
          >
            <div className="readmode-value u-truncate">
              {getReadModeValue()}
            </div>
            <Icon name="edit" className="edit-icon" />
          </Button>
        </>
      )}
    </>
  );
};

export default SettingForm;

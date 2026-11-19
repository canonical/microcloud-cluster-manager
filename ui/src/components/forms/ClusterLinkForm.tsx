import type { FC, FormEventHandler, ReactNode } from "react";
import {
  CheckboxInput,
  Form,
  Input,
  RadioInput,
} from "@canonical/react-components";
import type { FieldInputProps } from "formik";
import PanelSection from "components/PanelSection";
import ResourceIcon from "components/ResourceIcon";
import ClusterLinkName from "pages/clusters/ClusterLinkName";
import type { ClusterLink } from "types/cluster";
import { getLinkIconType, LINK_TYPES } from "util/clusterLinkTypes";

type Props = {
  sourceClusterName: string;
  type: ClusterLink["type"];
  name: string;
  descriptionField: FieldInputProps<string>;
  disabled: boolean;
  onSubmit: FormEventHandler<HTMLFormElement>;
  children?: ReactNode;
} & (
  | {
      mode: "create";
      onTypeChange: (type: ClusterLink["type"]) => void;
      targetSelector: ReactNode;
      addressControls?: ReactNode;
      targetClusterName?: never;
      applyDescriptionToBoth?: never;
      onApplyDescriptionToBothChange?: never;
    }
  | {
      mode: "edit";
      targetClusterName: string;
      applyDescriptionToBoth: boolean;
      onApplyDescriptionToBothChange: () => void;
      onTypeChange?: never;
      targetSelector?: never;
      addressControls?: never;
    }
);

const ClusterLinkForm: FC<Props> = ({
  sourceClusterName,
  type,
  name,
  descriptionField,
  disabled,
  onSubmit,
  children,
  mode,
  onTypeChange,
  targetSelector,
  addressControls,
  targetClusterName,
  applyDescriptionToBoth,
  onApplyDescriptionToBothChange,
}) => {
  const linkType = LINK_TYPES.find((linkType) => linkType.value === type);

  return (
    <Form onSubmit={onSubmit} className="form">
      {mode === "create" && <Input type="submit" hidden value="Hidden input" />}
      <PanelSection header="Link type" isFirst>
        {mode === "create" ? (
          <div className="cluster-link-types cluster-link-types--radio-buttons">
            {LINK_TYPES.map((linkType) => (
              <RadioInput
                key={linkType.value}
                name="type"
                label={linkType.label}
                help={linkType.description}
                checked={type === linkType.value}
                disabled={disabled}
                onChange={() => {
                  onTypeChange(linkType.value);
                }}
              />
            ))}
          </div>
        ) : (
          <div className="cluster-link-type-summary">
            <p className="link-type-label">{linkType?.label ?? type}</p>
            {linkType && (
              <p className="p-form-help-text">{linkType.description}</p>
            )}
          </div>
        )}
      </PanelSection>
      <PanelSection header="Details">
        <div className="cluster-link-endpoints">
          <Input
            type="text"
            label="Source"
            value={sourceClusterName}
            readOnly
          />
          <span className="cluster-link-endpoints__arrow" aria-hidden="true">
            <ResourceIcon type={getLinkIconType(type)} />
          </span>
          {mode === "create" ? (
            targetSelector
          ) : (
            <Input
              type="text"
              label="Target"
              value={targetClusterName}
              readOnly
            />
          )}
        </div>
        {mode === "create" && addressControls}
        <ClusterLinkName name={name} />
        <Input
          id="description"
          type="text"
          label="Description"
          className={mode === "edit" ? "u-no-margin--bottom" : undefined}
          disabled={disabled}
          {...descriptionField}
        />
        {mode === "edit" && type === "bidirectional" && (
          <CheckboxInput
            className="cluster-link-address-checkbox"
            label="Apply change in both source and target clusters"
            checked={applyDescriptionToBoth}
            disabled={disabled}
            onChange={onApplyDescriptionToBothChange}
          />
        )}
      </PanelSection>
      {children}
    </Form>
  );
};

export default ClusterLinkForm;

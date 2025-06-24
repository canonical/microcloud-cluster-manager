import { Button, Col, Icon, Label, Row } from "@canonical/react-components";
import type { FC, ReactNode } from "react";

interface Props {
  label: string;
  labelForId: string;
  value?: number;
  areAllValuesEqual: boolean;
  firstValue?: number;
  defaultValue: number;
  setValue: (value: number | undefined) => void;
  children: ReactNode;
}

const BulkConfigurePanelInput: FC<Props> = ({
  label,
  labelForId,
  value,
  areAllValuesEqual,
  firstValue,
  defaultValue,
  setValue,
  children,
}) => {
  return (
    <Row className="u-no-padding--left u-no-padding--right">
      <Col size={6}>
        <Label forId={labelForId}>
          <strong>{label}</strong>
        </Label>
      </Col>
      <Col size={6}>
        {value === undefined ? (
          <div className="u-flex">
            <div className="u-padding-top">
              {areAllValuesEqual ? (
                firstValue
              ) : (
                <span className="u-text--muted">Multiple values</span>
              )}
            </div>
            <Button
              appearance="base"
              type="button"
              hasIcon
              title="Set value"
              onClick={() => {
                setValue(areAllValuesEqual ? firstValue : defaultValue);
                setTimeout(() => {
                  document.getElementById(labelForId)?.focus();
                }, 50);
              }}
            >
              <Icon name="edit" />
            </Button>
          </div>
        ) : (
          <div className="u-flex">
            <div className="u-flex-grow">{children}</div>
            <Button
              appearance="base"
              type="button"
              hasIcon
              title="Undo change"
              onClick={() => {
                setValue(undefined);
              }}
            >
              <Icon name="close" />
            </Button>
          </div>
        )}
      </Col>
    </Row>
  );
};

export default BulkConfigurePanelInput;

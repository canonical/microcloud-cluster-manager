import { FC } from "react";
import { MultiMeterValue } from "types/cluster";

interface Props {
  values: MultiMeterValue[];
  text: string;
}

const MultiMeter: FC<Props> = ({ values, text }) => {
  function getPercentage(num: number): number {
    let cumulativeTotal = 0;

    for (const val of values) {
      cumulativeTotal += val.amount;
    }
    return (num / cumulativeTotal) * 100;
  }

  return (
    <>
      <div className="p-meter u-no-margin--bottom">
        {values.map((bar) => {
          return (
            <div
              key={bar.status}
              className={"meter-bar"}
              style={{
                width: `${getPercentage(bar.amount)}%`,
                backgroundColor: `${bar.color}`,
              }}
            />
          );
        })}
      </div>
      <div className="u-text--muted u-no-margin--bottom">{text}</div>
    </>
  );
};

export default MultiMeter;

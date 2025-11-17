import type { FC } from "react";

interface Props {
  percentage: number;
  text: string;
  type: "instances" | "cpu" | "disk" | "memory";
  containerClassname: string;
}

const Meter: FC<Props> = ({
  percentage,
  text,
  type,
  containerClassname,
}: Props) => {
  return (
    <div className={containerClassname}>
      <div className="p-meter u-no-margin--bottom">
        <div className={type} style={{ width: `${percentage}%` }} />
      </div>
      <div className="u-text--muted u-no-margin--bottom">{text}</div>
    </div>
  );
};

export default Meter;

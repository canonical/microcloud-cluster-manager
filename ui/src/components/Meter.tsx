import { FC } from "react";

interface Props {
  percentage: number;
  text: string;
  type: "instances" | "cpu" | "disk" | "memory";
}

const Meter: FC<Props> = ({ percentage, text, type }: Props) => {
  return (
    <>
      <div className="p-meter u-no-margin--bottom">
        <div className={type} style={{ width: `${percentage}%` }} />
      </div>
      <p className="p-text--small u-no-margin--bottom">{text}</p>
    </>
  );
};

export default Meter;

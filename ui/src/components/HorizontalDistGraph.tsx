import { FC } from "react";

type Props = {
  data: number[];
  title: string;
  color: string;
  width: number;
  height: number;
};

const X_GRID_LINE_RATIOS = [0.2, 0.4, 0.6, 0.8];

const HorizontalDistGraph: FC<Props> = ({
  data,
  title,
  color,
  width,
  height,
}) => {
  const maxDataValue = Math.max(...data);
  const barWidth = height / data.length;

  return (
    <div className="horizontal-dist-graph" style={{ maxWidth: `${width}px` }}>
      <p className="u-no-margin u-no-padding">{title}</p>
      <svg
        viewBox={`0 0 ${width} ${height}`}
        className="horizontal-dist-graph__chart"
      >
        {X_GRID_LINE_RATIOS.map((ratio) => (
          <line
            className="horizontal-dist-graph__grid-line"
            key={`${title}-line-${ratio}`}
            x1={ratio * width}
            x2={ratio * width}
            y1={0}
            y2={height}
          />
        ))}
        {data.map((dataPoint, index) => (
          <rect
            key={index}
            y={barWidth * index}
            x={0}
            // Add 0.5 to the height to prevent gaps between the bars
            height={barWidth + 0.5}
            width={(dataPoint / maxDataValue) * width}
            fill={color}
          />
        ))}
      </svg>
    </div>
  );
};

export default HorizontalDistGraph;

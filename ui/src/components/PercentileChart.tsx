import type { FC } from "react";

interface Props {
  data: number[];
  title: string;
  width: number;
  height: number;
  barClassName?: string;
}

const X_GRID_LINE_RATIOS = [0.2, 0.4, 0.6, 0.8];

const PercentileChart: FC<Props> = ({
  data,
  title,
  width,
  height,
  barClassName,
}) => {
  const barWidth = height / data.length;

  return (
    <div className="percentile-chart" style={{ maxWidth: `${width}px` }}>
      <p className="u-no-margin u-no-padding p-heading--5">{title}</p>
      <svg
        viewBox={`0 0 ${width} ${height}`}
        className="percentile-chart__chart"
      >
        {X_GRID_LINE_RATIOS.map((ratio) => (
          <line
            className="percentile-chart__grid-line"
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
            width={dataPoint * width}
            className={barClassName}
          />
        ))}
      </svg>
    </div>
  );
};

export default PercentileChart;

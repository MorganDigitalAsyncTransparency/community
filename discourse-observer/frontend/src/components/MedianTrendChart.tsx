// Spec: specs/dashboard/response-metrics.md
// Tests: manual (visual verification)

import {
  Line,
  LineChart,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from "recharts";
import type { MedianBucket } from "./medianTrendMetrics";
import { formatDuration } from "./topicFormatting";

const MILLISECONDS_PER_HOUR = 3_600_000;

function formatYAxisTick(hours: number): string {
  return formatDuration(hours * MILLISECONDS_PER_HOUR);
}

function formatTooltipValue(
  value: number | string | readonly (number | string)[] | undefined,
): string {
  if (typeof value !== "number") {
    return "\u2013";
  }
  return formatDuration(value * MILLISECONDS_PER_HOUR);
}

interface MedianTrendChartProps {
  data: MedianBucket[];
  color: string;
  name: string;
}

export function MedianTrendChart({ data, color, name }: MedianTrendChartProps) {
  return (
    <div className="median-trend-chart-wrapper">
      <ResponsiveContainer width="100%" height={250}>
        <LineChart data={data} margin={{ top: 5, right: 20, bottom: 5, left: 10 }}>
          <XAxis dataKey="label" tick={{ fontSize: 12 }} />
          <YAxis tickFormatter={formatYAxisTick} tick={{ fontSize: 12 }} />
          <Tooltip formatter={formatTooltipValue} />
          <Line
            type="monotone"
            dataKey="medianHours"
            name={name}
            stroke={color}
            connectNulls={false}
            dot={{ r: 3 }}
            activeDot={{ r: 5 }}
          />
        </LineChart>
      </ResponsiveContainer>
    </div>
  );
}

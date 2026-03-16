// Spec: specs/dashboard/response-time-trends.md
// Tests: tests/dashboard/response-time-trends.unit.test.ts

import {
  Legend,
  Line,
  LineChart,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from "recharts";
import type { TrendChartPoint } from "./trendMetrics";
import { formatDuration } from "./topicFormatting";

const FIRST_REPLY_COLOR = "#8884d8";
const RESOLUTION_COLOR = "#82ca9d";
const MILLISECONDS_PER_HOUR = 3_600_000;

function formatYAxisTick(hours: number): string {
  return formatDuration(hours * MILLISECONDS_PER_HOUR);
}

function formatTooltipValue(
  value: number | string | readonly (number | string)[] | undefined,
): string {
  if (typeof value !== "number") {
    return "–";
  }
  return formatDuration(value * MILLISECONDS_PER_HOUR);
}

interface ResponseTimeTrendChartProps {
  data: TrendChartPoint[];
}

export function ResponseTimeTrendChart({ data }: ResponseTimeTrendChartProps) {
  return (
    <div className="trends-chart-wrapper">
      <ResponsiveContainer width="100%" height={300}>
        <LineChart data={data} margin={{ top: 5, right: 20, bottom: 5, left: 10 }}>
          <XAxis dataKey="weekLabel" tick={{ fontSize: 12 }} />
          <YAxis tickFormatter={formatYAxisTick} tick={{ fontSize: 12 }} />
          <Tooltip formatter={formatTooltipValue} />
          <Legend />
          <Line
            type="monotone"
            dataKey="medianFirstReplyHours"
            name="Median first reply"
            stroke={FIRST_REPLY_COLOR}
            connectNulls={false}
            dot={{ r: 3 }}
            activeDot={{ r: 5 }}
          />
          <Line
            type="monotone"
            dataKey="medianResolutionHours"
            name="Median resolution"
            stroke={RESOLUTION_COLOR}
            connectNulls={false}
            dot={{ r: 3 }}
            activeDot={{ r: 5 }}
          />
        </LineChart>
      </ResponsiveContainer>
    </div>
  );
}

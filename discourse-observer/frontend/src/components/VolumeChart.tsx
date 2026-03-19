// Spec: specs/dashboard/response-metrics.md
// Tests: manual (visual verification)

import {
  Legend,
  Line,
  LineChart,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from "recharts";
import type { VolumeBucket } from "../api/types";
import {
  CHART_COLOR_3,
  CHART_COLOR_4,
  CHART_COLOR_5,
  CHART_COLOR_6,
} from "./themeColors";

interface VolumeChartProps {
  data: VolumeBucket[];
}

export function VolumeChart({ data }: VolumeChartProps) {
  return (
    <div className="volume-chart-wrapper">
      <ResponsiveContainer width="100%" height={300}>
        <LineChart data={data} margin={{ top: 5, right: 20, bottom: 5, left: 10 }}>
          <XAxis dataKey="label" tick={{ fontSize: 12 }} />
          <YAxis allowDecimals={false} tick={{ fontSize: 12 }} />
          <Tooltip />
          <Legend />
          <Line
            type="monotone"
            dataKey="created"
            name="Topics created"
            stroke={CHART_COLOR_3}
            dot={{ r: 3 }}
            activeDot={{ r: 5 }}
          />
          <Line
            type="monotone"
            dataKey="accepted"
            name="Accepted answer"
            stroke={CHART_COLOR_6}
            dot={{ r: 3 }}
            activeDot={{ r: 5 }}
          />
          <Line
            type="monotone"
            dataKey="closed"
            name="Topics closed"
            stroke={CHART_COLOR_5}
            dot={{ r: 3 }}
            activeDot={{ r: 5 }}
          />
          <Line
            type="monotone"
            dataKey="open"
            name="Currently open"
            stroke={CHART_COLOR_4}
            dot={{ r: 3 }}
            activeDot={{ r: 5 }}
          />
        </LineChart>
      </ResponsiveContainer>
    </div>
  );
}

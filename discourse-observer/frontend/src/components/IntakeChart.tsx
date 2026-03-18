// Spec: specs/dashboard/topic-intake.md
// Tests: tests/dashboard/topic-intake.unit.test.ts

import {
  Line,
  LineChart,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from "recharts";
import type { IntakeBucket } from "./intakeMetrics";
import { getThemeColor } from "./useThemeColors";

interface IntakeChartProps {
  data: IntakeBucket[];
}

export function IntakeChart({ data }: IntakeChartProps) {
  const lineColor = getThemeColor("--color-chart-3");

  return (
    <div className="intake-chart-wrapper">
      <ResponsiveContainer width="100%" height={300}>
        <LineChart data={data} margin={{ top: 5, right: 20, bottom: 5, left: 10 }}>
          <XAxis dataKey="label" tick={{ fontSize: 12 }} />
          <YAxis allowDecimals={false} tick={{ fontSize: 12 }} />
          <Tooltip />
          <Line
            type="monotone"
            dataKey="count"
            name="Topics"
            stroke={lineColor}
            dot={{ r: 3 }}
          />
        </LineChart>
      </ResponsiveContainer>
    </div>
  );
}

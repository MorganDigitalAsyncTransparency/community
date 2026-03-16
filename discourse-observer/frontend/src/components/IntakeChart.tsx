// Spec: specs/dashboard/topic-intake.md
// Tests: tests/dashboard/topic-intake.unit.test.ts

import {
  Bar,
  BarChart,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from "recharts";
import type { IntakeBucket } from "./intakeMetrics";

const BAR_COLOR = "#5b8ff9";

interface IntakeChartProps {
  data: IntakeBucket[];
}

export function IntakeChart({ data }: IntakeChartProps) {
  return (
    <div className="intake-chart-wrapper">
      <ResponsiveContainer width="100%" height={300}>
        <BarChart data={data} margin={{ top: 5, right: 20, bottom: 5, left: 10 }}>
          <XAxis dataKey="label" tick={{ fontSize: 12 }} />
          <YAxis allowDecimals={false} tick={{ fontSize: 12 }} />
          <Tooltip />
          <Bar dataKey="count" name="Topics" fill={BAR_COLOR} />
        </BarChart>
      </ResponsiveContainer>
    </div>
  );
}

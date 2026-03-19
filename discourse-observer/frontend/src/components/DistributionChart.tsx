// Spec: specs/dashboard/response-time-distribution.md
// Tests: backend/api/contract_test.go

import {
  Bar,
  BarChart,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from "recharts";
import type { DistributionBucket } from "../api/types";

interface DistributionChartProps {
  data: DistributionBucket[];
  color: string;
  name: string;
}

export function DistributionChart({ data, color, name }: DistributionChartProps) {
  return (
    <div className="rd-chart-wrapper">
      <ResponsiveContainer width="100%" height={300}>
        <BarChart data={data} margin={{ top: 5, right: 20, bottom: 5, left: 10 }}>
          <XAxis dataKey="label" tick={{ fontSize: 12 }} />
          <YAxis allowDecimals={false} tick={{ fontSize: 12 }} />
          <Tooltip />
          <Bar dataKey="count" name={name} fill={color} />
        </BarChart>
      </ResponsiveContainer>
    </div>
  );
}

"use client";

import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ReferenceLine,
  ResponsiveContainer,
} from "recharts";

interface ChartDataPoint {
  time: string;
  min: number;
  avg: number;
  max: number;
}

interface PriceChartGraphProps {
  data: ChartDataPoint[];
  alertPrice: number;
}

export default function PriceChartGraph({
  data,
  alertPrice,
}: PriceChartGraphProps) {
  if (data.length === 0) {
    return (
      <p className="text-[var(--text-secondary)] text-center py-12">
        No price data yet for this period.
      </p>
    );
  }

  return (
    <ResponsiveContainer width="100%" height={350}>
      <LineChart data={data}>
        <CartesianGrid strokeDasharray="3 3" stroke="rgba(255,255,255,0.06)" />
        <XAxis
          dataKey="time"
          tick={{ fontSize: 11, fill: "#64748b" }}
          axisLine={{ stroke: "rgba(255,255,255,0.1)" }}
          tickLine={{ stroke: "rgba(255,255,255,0.1)" }}
          interval="preserveStartEnd"
        />
        <YAxis
          tick={{ fontSize: 11, fill: "#64748b" }}
          axisLine={{ stroke: "rgba(255,255,255,0.1)" }}
          tickLine={{ stroke: "rgba(255,255,255,0.1)" }}
          tickFormatter={(v) => `$${v}`}
        />
        <Tooltip
          formatter={(value) => [`$${Number(value).toFixed(2)}`, ""]}
          contentStyle={{
            backgroundColor: "#1e293b",
            border: "1px solid rgba(255,255,255,0.1)",
            borderRadius: "8px",
            color: "#f1f5f9",
          }}
          itemStyle={{ color: "#94a3b8" }}
          labelStyle={{ color: "#f1f5f9" }}
        />
        <ReferenceLine
          y={alertPrice}
          stroke="#ef4444"
          strokeDasharray="5 5"
          label={{
            value: `Alert $${alertPrice}`,
            position: "right",
            fill: "#f87171",
            fontSize: 12,
          }}
        />
        <Line
          type="monotone"
          dataKey="min"
          stroke="#06b6d4"
          strokeWidth={2}
          dot={false}
          name="Min Price"
        />
        <Line
          type="monotone"
          dataKey="avg"
          stroke="#f59e0b"
          strokeWidth={2}
          dot={false}
          name="Avg Price"
        />
        <Line
          type="monotone"
          dataKey="max"
          stroke="rgba(255,255,255,0.2)"
          strokeWidth={1}
          dot={false}
          name="Max Price"
          strokeDasharray="3 3"
        />
      </LineChart>
    </ResponsiveContainer>
  );
}

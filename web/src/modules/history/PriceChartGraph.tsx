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
        <CartesianGrid strokeDasharray="3 3" stroke="var(--border-default)" />
        <XAxis
          dataKey="time"
          tick={{ fontSize: 11 }}
          interval="preserveStartEnd"
        />
        <YAxis tick={{ fontSize: 11 }} tickFormatter={(v) => `$${v}`} />
        <Tooltip
          formatter={(value) => [`$${Number(value).toFixed(2)}`, ""]}
        />
        <ReferenceLine
          y={alertPrice}
          stroke="#ef4444"
          strokeDasharray="5 5"
          label={{
            value: `Alert $${alertPrice}`,
            position: "right",
            fill: "#ef4444",
            fontSize: 12,
          }}
        />
        <Line
          type="monotone"
          dataKey="min"
          stroke="#059669"
          strokeWidth={2}
          dot={false}
          name="Min Price"
        />
        <Line
          type="monotone"
          dataKey="avg"
          stroke="#2563eb"
          strokeWidth={2}
          dot={false}
          name="Avg Price"
        />
        <Line
          type="monotone"
          dataKey="max"
          stroke="#f97316"
          strokeWidth={1}
          dot={false}
          name="Max Price"
          strokeDasharray="3 3"
        />
      </LineChart>
    </ResponsiveContainer>
  );
}

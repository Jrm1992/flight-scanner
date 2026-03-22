"use client";

import type { PriceInsights } from "@/lib/types";
import {
  AreaChart,
  Area,
  XAxis,
  YAxis,
  Tooltip,
  ResponsiveContainer,
} from "recharts";

const currencySymbols: Record<string, string> = {
  BRL: "R$",
  USD: "$",
  EUR: "\u20AC",
  GBP: "\u00A3",
  ARS: "ARS$",
  CLP: "CLP$",
  COP: "COP$",
};

const levelColors: Record<string, { bg: string; text: string; label: string }> = {
  low: { bg: "bg-emerald-500/15", text: "text-emerald-400", label: "Low" },
  typical: { bg: "bg-amber-500/15", text: "text-amber-400", label: "Typical" },
  high: { bg: "bg-red-500/15", text: "text-red-400", label: "High" },
};

interface PriceInsightsCardProps {
  insights: PriceInsights;
  currency: string;
}

export default function PriceInsightsCard({
  insights,
  currency,
}: PriceInsightsCardProps) {
  const sym = currencySymbols[currency] || currency;
  const level = levelColors[insights.price_level] || levelColors.typical;

  const chartData = insights.price_history?.map(([ts, price]) => ({
    date: new Date(ts * 1000).toLocaleDateString("en", { month: "short", day: "numeric" }),
    price,
  })) ?? [];

  const [low, high] = insights.typical_price_range ?? [0, 0];

  return (
    <div className="rounded-lg border border-border bg-white/5 backdrop-blur-xl p-4 mb-4">
      <div className="flex flex-wrap items-center gap-4 mb-3">
        <span className={`px-2 py-0.5 rounded text-xs font-semibold ${level.bg} ${level.text}`}>
          {level.label}
        </span>
        <span className="text-sm text-foreground">
          Lowest: <span className="font-semibold text-emerald-400">{sym} {insights.lowest_price}</span>
        </span>
        {low > 0 && high > 0 && (
          <span className="text-sm text-muted-foreground">
            Typical range: {sym} {low} &ndash; {sym} {high}
          </span>
        )}
      </div>

      {chartData.length > 2 && (
        <ResponsiveContainer width="100%" height={120}>
          <AreaChart data={chartData}>
            <defs>
              <linearGradient id="priceGrad" x1="0" y1="0" x2="0" y2="1">
                <stop offset="0%" stopColor="#06b6d4" stopOpacity={0.3} />
                <stop offset="100%" stopColor="#06b6d4" stopOpacity={0} />
              </linearGradient>
            </defs>
            <XAxis
              dataKey="date"
              tick={{ fontSize: 10, fill: "#64748b" }}
              axisLine={false}
              tickLine={false}
              interval="preserveStartEnd"
            />
            <YAxis
              tick={{ fontSize: 10, fill: "#64748b" }}
              axisLine={false}
              tickLine={false}
              tickFormatter={(v) => `${sym}${v}`}
              width={60}
            />
            <Tooltip
              formatter={(value: number) => [`${sym} ${value}`, "Price"]}
              contentStyle={{
                backgroundColor: "#1e293b",
                border: "1px solid rgba(255,255,255,0.1)",
                borderRadius: "8px",
                color: "#f1f5f9",
                fontSize: "12px",
              }}
            />
            <Area
              type="monotone"
              dataKey="price"
              stroke="#06b6d4"
              strokeWidth={2}
              fill="url(#priceGrad)"
            />
          </AreaChart>
        </ResponsiveContainer>
      )}
    </div>
  );
}

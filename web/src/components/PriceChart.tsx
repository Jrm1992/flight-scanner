"use client";

import { useState, useEffect, useCallback } from "react";
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
import { getHistory } from "@/lib/api";
import type { Route, PriceHistory, PriceStats } from "@/lib/types";

interface Props {
  route: Route;
  onClose: () => void;
}

const PERIODS = [
  { label: "7d", value: 7 },
  { label: "30d", value: 30 },
  { label: "90d", value: 90 },
];

export default function PriceChart({ route, onClose }: Props) {
  const [history, setHistory] = useState<PriceHistory[]>([]);
  const [stats, setStats] = useState<PriceStats | null>(null);
  const [days, setDays] = useState(30);
  const [loading, setLoading] = useState(true);

  const loadHistory = useCallback(async () => {
    setLoading(true);
    try {
      const data = await getHistory(route.id, days);
      setHistory(data.history || []);
      setStats(data.stats);
    } catch {
      setHistory([]);
    } finally {
      setLoading(false);
    }
  }, [route.id, days]);

  useEffect(() => {
    loadHistory();
  }, [loadHistory]);

  const chartData = history.map((h) => ({
    time: new Date(h.checked_at).toLocaleString("en-US", {
      month: "short",
      day: "numeric",
      hour: "2-digit",
      minute: "2-digit",
    }),
    min: h.min_price,
    avg: h.avg_price,
    max: h.max_price,
  }));

  return (
    <div>
      <div className="flex items-center justify-between mb-4">
        <div>
          <h2 className="text-xl font-semibold">
            {route.origin} → {route.destination}
          </h2>
          <p className="text-sm text-gray-500">Price History</p>
        </div>
        <div className="flex items-center gap-2">
          {PERIODS.map((p) => (
            <button
              key={p.value}
              onClick={() => setDays(p.value)}
              className={`px-3 py-1 text-sm rounded ${
                days === p.value
                  ? "bg-blue-600 text-white"
                  : "bg-gray-100 text-gray-700 hover:bg-gray-200"
              }`}
            >
              {p.label}
            </button>
          ))}
          <button
            onClick={onClose}
            className="ml-2 text-gray-500 hover:text-gray-700 text-sm px-3 py-1 border border-gray-200 rounded"
          >
            Close
          </button>
        </div>
      </div>

      {stats && (
        <div className="grid grid-cols-3 gap-4 mb-6">
          <div className="bg-green-50 rounded-lg p-3 text-center">
            <p className="text-xs text-gray-500">Min Price</p>
            <p className="text-lg font-bold text-green-700">
              ${stats.min_price.toFixed(0)}
            </p>
          </div>
          <div className="bg-blue-50 rounded-lg p-3 text-center">
            <p className="text-xs text-gray-500">Avg Price</p>
            <p className="text-lg font-bold text-blue-700">
              ${stats.avg_price.toFixed(0)}
            </p>
          </div>
          <div className="bg-red-50 rounded-lg p-3 text-center">
            <p className="text-xs text-gray-500">Max Price</p>
            <p className="text-lg font-bold text-red-700">
              ${stats.max_price.toFixed(0)}
            </p>
          </div>
        </div>
      )}

      {loading ? (
        <p className="text-gray-500 text-center py-12">Loading chart...</p>
      ) : chartData.length === 0 ? (
        <p className="text-gray-500 text-center py-12">
          No price data yet for this period.
        </p>
      ) : (
        <ResponsiveContainer width="100%" height={350}>
          <LineChart data={chartData}>
            <CartesianGrid strokeDasharray="3 3" stroke="#e5e7eb" />
            <XAxis
              dataKey="time"
              tick={{ fontSize: 11 }}
              interval="preserveStartEnd"
            />
            <YAxis
              tick={{ fontSize: 11 }}
              tickFormatter={(v) => `$${v}`}
            />
            <Tooltip
              formatter={(value) => [`$${Number(value).toFixed(2)}`, ""]}
            />
            <ReferenceLine
              y={route.alert_price}
              stroke="#ef4444"
              strokeDasharray="5 5"
              label={{
                value: `Alert $${route.alert_price}`,
                position: "right",
                fill: "#ef4444",
                fontSize: 12,
              }}
            />
            <Line
              type="monotone"
              dataKey="min"
              stroke="#22c55e"
              strokeWidth={2}
              dot={false}
              name="Min Price"
            />
            <Line
              type="monotone"
              dataKey="avg"
              stroke="#3b82f6"
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
      )}
    </div>
  );
}

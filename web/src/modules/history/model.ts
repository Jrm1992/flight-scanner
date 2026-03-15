import { useState, useCallback, useEffect } from "react";
import { getHistory, getExportUrl } from "@/lib/api";
import type { PriceHistory, PriceStats } from "@/lib/types";

export const PERIODS = [
  { label: "7d", value: 7 },
  { label: "30d", value: 30 },
  { label: "90d", value: 90 },
];

export function useHistoryModel(routeId: string) {
  const [days, setDays] = useState(30);
  const [history, setHistory] = useState<PriceHistory[]>([]);
  const [stats, setStats] = useState<PriceStats | null>(null);
  const [loading, setLoading] = useState(true);

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const data = await getHistory(routeId, days);
      setHistory(data.history || []);
      setStats(data.stats);
    } catch {
      setHistory([]);
    } finally {
      setLoading(false);
    }
  }, [routeId, days]);

  useEffect(() => {
    load();
  }, [load]);

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

  const exportCsvUrl = getExportUrl(routeId, days, "csv");
  const exportJsonUrl = getExportUrl(routeId, days, "json");

  return {
    days,
    setDays,
    periods: PERIODS,
    chartData,
    stats,
    loading,
    exportCsvUrl,
    exportJsonUrl,
  };
}

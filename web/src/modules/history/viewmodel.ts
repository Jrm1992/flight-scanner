import { useState } from "react";
import { useHistoryModel } from "./model";
import { getExportUrl } from "@/lib/api";
import type { Route } from "@/lib/types";

export const PERIODS = [
  { label: "7d", value: 7 },
  { label: "30d", value: 30 },
  { label: "90d", value: 90 },
];

export function useHistoryViewModel(route: Route) {
  const [days, setDays] = useState(30);
  const model = useHistoryModel(route.id, days);

  const chartData = model.history.map((h) => ({
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

  const exportCsvUrl = getExportUrl(route.id, days, "csv");
  const exportJsonUrl = getExportUrl(route.id, days, "json");

  return {
    days,
    setDays,
    periods: PERIODS,
    chartData,
    stats: model.stats,
    loading: model.loading,
    exportCsvUrl,
    exportJsonUrl,
  };
}

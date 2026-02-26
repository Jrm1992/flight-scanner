import { useState, useCallback, useEffect } from "react";
import { getHistory } from "@/lib/api";
import type { PriceHistory, PriceStats } from "@/lib/types";

export function useHistoryModel(routeId: string, days: number) {
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

  return { history, stats, loading };
}

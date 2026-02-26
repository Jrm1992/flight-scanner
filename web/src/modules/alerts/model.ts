import { useState, useCallback, useEffect } from "react";
import { getAlerts, markAlertRead } from "@/lib/api";
import type { Alert } from "@/lib/types";

export function useAlertsModel() {
  const [alerts, setAlerts] = useState<Alert[]>([]);
  const [loading, setLoading] = useState(true);

  const loadAlerts = useCallback(async () => {
    try {
      const data = await getAlerts();
      setAlerts(data);
    } catch {
      setAlerts([]);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    loadAlerts();
  }, [loadAlerts]);

  const markRead = useCallback(
    async (id: string) => {
      await markAlertRead(id);
      await loadAlerts();
    },
    [loadAlerts]
  );

  return { alerts, loading, loadAlerts, markRead };
}

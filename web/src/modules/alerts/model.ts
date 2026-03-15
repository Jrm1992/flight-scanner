import { useState, useCallback, useEffect, useMemo } from "react";
import { getAlerts, markAlertRead, getRoutes } from "@/lib/api";
import type { Alert, Route } from "@/lib/types";

export function useAlertsModel() {
  const [alerts, setAlerts] = useState<Alert[]>([]);
  const [routes, setRoutes] = useState<Route[]>([]);
  const [loading, setLoading] = useState(true);
  const [filter, setFilter] = useState<"all" | "unread" | "read">("all");
  const [routeFilter, setRouteFilter] = useState("");

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

  const loadRoutes = useCallback(async () => {
    try {
      const data = await getRoutes();
      setRoutes(data);
    } catch {
      setRoutes([]);
    }
  }, []);

  useEffect(() => {
    loadAlerts();
    loadRoutes();
  }, [loadAlerts, loadRoutes]);

  const routeMap = useMemo(() => {
    const map = new Map<string, Route>();
    routes.forEach((r) => map.set(r.id, r));
    return map;
  }, [routes]);

  const filteredAlerts = alerts.filter((a) => {
    if (filter === "unread" && a.notified) return false;
    if (filter === "read" && !a.notified) return false;
    if (routeFilter && a.route_id !== routeFilter) return false;
    return true;
  });

  async function handleMarkRead(id: string) {
    try {
      await markAlertRead(id);
      await loadAlerts();
    } catch {
      // ignore
    }
  }

  return {
    alerts: filteredAlerts,
    routes,
    routeMap,
    loading,
    filter,
    setFilter,
    routeFilter,
    setRouteFilter,
    handleMarkRead,
  };
}

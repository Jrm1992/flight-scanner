import { useState, useMemo } from "react";
import { useAlertsModel } from "./model";
import { useRoutesModel } from "@/modules/routes/model";
import type { Route } from "@/lib/types";

export function useAlertsViewModel() {
  const alertsModel = useAlertsModel();
  const routesModel = useRoutesModel();

  const [filter, setFilter] = useState<"all" | "unread" | "read">("all");
  const [routeFilter, setRouteFilter] = useState("");

  const routeMap = useMemo(() => {
    const map = new Map<string, Route>();
    routesModel.routes.forEach((r) => map.set(r.id, r));
    return map;
  }, [routesModel.routes]);

  const filteredAlerts = alertsModel.alerts.filter((a) => {
    if (filter === "unread" && a.notified) return false;
    if (filter === "read" && !a.notified) return false;
    if (routeFilter && a.route_id !== routeFilter) return false;
    return true;
  });

  async function handleMarkRead(id: string) {
    try {
      await alertsModel.markRead(id);
    } catch {
      // ignore
    }
  }

  return {
    alerts: filteredAlerts,
    routes: routesModel.routes,
    routeMap,
    loading: alertsModel.loading,
    filter,
    setFilter,
    routeFilter,
    setRouteFilter,
    handleMarkRead,
  };
}

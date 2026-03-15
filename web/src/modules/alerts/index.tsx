"use client";

import { useAlertsModel } from "./model";
import AlertsView from "./view";

export default function Alerts() {
  const model = useAlertsModel();

  return (
    <AlertsView
      alerts={model.alerts}
      routes={model.routes}
      routeMap={model.routeMap}
      loading={model.loading}
      filter={model.filter}
      onFilterChange={(v) => model.setFilter(v as "all" | "unread" | "read")}
      routeFilter={model.routeFilter}
      onRouteFilterChange={model.setRouteFilter}
      onMarkRead={model.handleMarkRead}
    />
  );
}
